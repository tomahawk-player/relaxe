/* === This file is part of Tomahawk Player - <http://tomahawk-player.org> ===
 *
 *   Copyright 2013, Teo Mrnjavac <teo@kde.org>
 *
 *   Tomahawk is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   Tomahawk is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with Tomahawk. If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"github.com/teo/relaxe/common"
	"github.com/teo/relaxe/makeaxe/bundle"
	"github.com/teo/relaxe/makeaxe/util"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	programName        = "makeaxe"
	programDescription = "the Tomahawk resolver bundle creator"
	programVersion     = "0.1"
)

var (
	all     bool
	release bool
	force   bool
	help    bool
	verbose bool
	relaxe  bool
)

func usage() {
	fmt.Printf("*** %v %v - %v ***\n\n", programName, programVersion, programDescription)
	fmt.Println("Usage: ./makeaxe [OPTIONS] SOURCE [DESTINATION|CONFIG]")
	fmt.Println("OPTIONS")
	flag.VisitAll(func(f *flag.Flag) {
		if len(f.Name) < 2 {
			return
		}
		fmt.Printf("\t%v\n", f.Usage)
	})
	fmt.Println("ARGUMENTS")
	fmt.Println("\tSOURCE\t\tMandatory, the path of the unpackaged base directory. " +
		"\n\t\t\tWhen building a single axe, this should be the parent of the directory that contains a metadata.json file. " +
		"\n\t\t\tIf building all resolvers (--all, -a) this should be the parent directory of all the resolvers.")

	fmt.Println("\tDESTINATION\tOptional, the path of the directory where newly built bundles (axes) should be placed. " +
		"\n\t\t\tIf unset, it is the same as the source directory. Not used when publishing to Relaxe (--relaxe, -x).")

	fmt.Println("\tCONFIG\t\tOnly when publishing to Relaxe (--relaxe, -x), the path of the Relaxe configuration file.")
}

func die(message string) {
	fmt.Println(message)
	fmt.Println("See ./makeaxe --help for usage information.")
	os.Exit(2)
}

func init() {
	const (
		flagAllUsage     = "--all, -a\tbuild all the resolvers in the SOURCE path's subdirectories"
		flagReleaseUsage = "--release, -r\tskip trying to add the git revision hash to a bundle"
		flagForceUsage   = "--force, -f\tbuild a bundle and overwrite even if the destination directory already contains a bundle of the same name and version"
		flagHelpUsage    = "--help, -h\tthis help message"
		flagVerbose      = "--verbose, -v\tshow verbose output"
		flagRelaxeUsage  = "--relaxe, -x\tpublish resolvers on a Relaxe instance with the given config file, implies --release and ignores --force and DESTINATION"
	)
	flag.BoolVar(&all, "all", false, flagAllUsage)
	flag.BoolVar(&all, "a", false, flagAllUsage+" (shorthand)")
	flag.BoolVar(&release, "release", false, flagReleaseUsage)
	flag.BoolVar(&release, "r", false, flagReleaseUsage+" (shorthand)")
	flag.BoolVar(&force, "force", false, flagForceUsage)
	flag.BoolVar(&force, "f", false, flagForceUsage)
	flag.BoolVar(&help, "help", false, flagHelpUsage)
	flag.BoolVar(&help, "h", false, flagHelpUsage)
	flag.BoolVar(&verbose, "verbose", false, flagVerbose)
	flag.BoolVar(&verbose, "v", false, flagVerbose)
	flag.BoolVar(&relaxe, "relaxe", false, flagRelaxeUsage)
	flag.BoolVar(&relaxe, "x", false, flagRelaxeUsage)
}

func preparePaths(inputPath string) []string {
	inputList := []string{}
	if all {
		contents, err := ioutil.ReadDir(inputPath)
		if err != nil {
			die(err.Error())
		}
		for _, entry := range contents {
			if !entry.IsDir() {
				continue
			}
			realInputPath := path.Join(inputPath, entry.Name())
			metadataPath := path.Join(realInputPath, "content", "metadata.json")
			ex, err := util.ExistsFile(metadataPath)
			if !ex || err != nil {
				log.Printf("%v does not seem to be an axe directory, skipping.\n", entry.Name())
				continue
			}
			inputList = append(inputList, realInputPath)
		}

	} else {
		inputList = append(inputList, inputPath)
	}
	return inputList
}

func buildToRelaxe(inputList []string, relaxeConfig common.RelaxeConfig) string {
	if !relaxe {
		die("Error: cannot push to Relaxe in directory mode.")
	}

	// Try to connect to the MongoDB instance first, bail out if we can't
	session, err := mgo.Dial(relaxeConfig.Database.ConnectionString)
	if err != nil {
		die("Error: cannot connect to Relaxe database. Reason: " + err.Error())
	}
	c := session.DB("relaxe").C("axes")

	log.Println("Connected to Relaxe MongoDB instance, collection: " + c.FullName)

	built := []string{}
	errors := []string{}
	skipped := []string{}

	outputPath := relaxeConfig.CacheDirectory
	for _, inputDirPath := range inputList {
		b, err := bundle.LoadBundle(inputDirPath)
		if err != nil {
			log.Printf("Warning: could not load bundle from directory %v.\n", inputDirPath)
			skipped = append(skipped, path.Base(inputDirPath))
			continue
		}

		count, err := c.Find(bson.M{"pluginname": b.Metadata.PluginName, "version": b.Metadata.Version}).Count()

		if err != nil {
			log.Printf("Warning: Relaxe database error. %v\n", err.Error())
			errors = append(errors, path.Base(inputDirPath))
			continue
		}
		if count != 0 { //if Relaxe already has axes of the same pluginName and version
			log.Printf("Warning: axe %v-%v is already published on Relaxe, skipping.\n", b.Metadata.PluginName, b.Metadata.Version)
			skipped = append(skipped, path.Base(inputDirPath))
			continue
		}

		u, err := uuid.NewV4()
		axeUuid := u.String()

		b.Metadata.AxeId = axeUuid

		outputFilePath, err := b.CreatePackage(outputPath, true /*release*/, false /*force*/)
		if err != nil {
			log.Printf("Warning: could not build axe for directory %v.\n", path.Base(inputDirPath))
			errors = append(errors, path.Base(inputDirPath))
			continue
		}
		log.Printf("* Created axe in %v.\n", outputFilePath)

		mrshld, _ := json.MarshalIndent(b.Metadata, "", "  ")
		log.Println("* Pushing to Relaxe:\n" + string(mrshld))
		c.Insert(b.Metadata)

		built = append(built, "UUID:"+axeUuid+"\t"+b.Metadata.PluginName+"-"+b.Metadata.Version)
	}

	preamble := fmt.Sprintf("Relaxe instance at %v; pushing to cache directory: %v\n", strings.Join(session.LiveServers(), ", "), outputPath)
	return makeSummary(preamble, built, errors, skipped)
}

func buildToDirectory(inputList []string, outputPath string) string {
	if relaxe {
		die("Error: cannot build to directory in Relaxe mode.")
	}

	built := []string{}
	errors := []string{}
	skipped := []string{}

	for _, inputDirPath := range inputList {
		b, err := bundle.LoadBundle(inputDirPath)
		if err != nil {
			log.Printf("Warning: could not load bundle from directory %v.\n", inputDirPath)
			skipped = append(skipped, path.Base(inputDirPath))
			continue
		}
		outputFilePath, err := b.CreatePackage(outputPath, release, force)
		if err != nil {
			log.Printf("Warning: could not build axe for directory %v. %v\n", path.Base(inputDirPath), err.Error())
			if outputFilePath != "" { //means we are not creating just because the axe already exists
				skipped = append(skipped, path.Base(inputDirPath))
			} else {
				errors = append(errors, path.Base(inputDirPath))
			}
			continue
		}
		log.Printf("* Created axe in %v.\n", outputFilePath)
		built = append(built, path.Base(outputFilePath))
	}

	return makeSummary("Output directory: "+outputPath+"\n", built, errors, skipped)
}

func makeSummary(preamble string, built []string, errors []string, skipped []string) string {
	var (
		builtText   string
		errorsText  string
		skippedText string
	)
	if len(built) == 0 {
		builtText = fmt.Sprint("No axes built\n")
	} else {
		builtText = fmt.Sprintf("Axes built: %v\n"+
			"    * %v\n", len(built), strings.Join(built, "\n    * "))
	}

	if len(errors) != 0 {
		errorsText = fmt.Sprintf("Build errors: %v\n"+
			"    * %v\n", len(errors), strings.Join(errors, "\n    * "))
	}

	if len(skipped) != 0 {
		skippedText = fmt.Sprintf("Directories skipped: %v\n"+
			"    * %v\n", len(skipped), strings.Join(skipped, "\n    * "))
	}

	summary := fmt.Sprintf("*** makeaxe Summary ***\n\n%v%v%v%v", preamble, builtText, errorsText, skippedText)
	return summary
}

func main() {
	flag.Parse()

	if help {
		usage()
		return
	}

	if !verbose {
		log.SetOutput(ioutil.Discard)
	}

	if len(flag.Args()) == 0 {
		die("Error: a source directory must be specified.")
	}

	if len(flag.Args()) > 2 {
		die("Error: too many arguments.")
	}

	// Prepare input directory path(s)
	inputPath, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		die("Error: bad source directory path.")
	}
	if ex, err := util.ExistsDir(inputPath); !ex || err != nil {
		die("Error: bad source directory path.")
	}

	inputList := preparePaths(inputPath)

	var summary string

	// Prepare output directory path and build
	if relaxe {
		if len(flag.Args()) != 2 {
			die("Error: source or Relaxe configuration file path missing.")
		}
		configFilePath, err := filepath.Abs(flag.Arg(1))
		if err != nil {
			die("Error: bad Relaxe configuration file path.")
		}
		if ex, err := util.ExistsFile(configFilePath); !ex || err != nil {
			die("Error: bad Relaxe configuration file path.")
		}

		config, err := common.LoadConfig(configFilePath)
		if err != nil {
			die(err.Error())
		}

		summary = buildToRelaxe(inputList, *config)

	} else {
		var outputPath string

		if len(flag.Args()) == 1 {
			outputPath = inputPath
		} else { //len is 2
			outputPath, err = filepath.Abs(flag.Arg(1))
			if err != nil {
				die("Error: bad destination directory path.")
			}
			if ex, err := util.ExistsDir(outputPath); !ex || err != nil {
				die("Error: bad destination directory path.")
			}
		}

		summary = buildToDirectory(inputList, outputPath)
	}

	fmt.Printf(summary)
}
