/* === This file is part of Relaxe - <https://github.com/teo/relaxe> ===
 *
 *   Copyright 2013, Teo Mrnjavac <teo@kde.org>
 *
 *   Relaxe is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   Relaxe is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with Relaxe. If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"flag"
	"fmt"
	"github.com/teo/relaxe/common"
	"github.com/teo/relaxe/common/util"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

	flag.Usage = usage
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
