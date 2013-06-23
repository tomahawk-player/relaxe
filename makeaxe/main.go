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
	"flag"
	"fmt"
	"github.com/teo/relaxe/makeaxe/bundle"
	"os"
	"path/filepath"
)

const (
	programName        = "makeaxe"
	programDescription = "the Tomahawk resolver bundle creator"
	programVersion     = "0.1"
	bundleVersion      = "1"
)

var (
	all     bool
	release bool
	force   bool
	help    bool
	ver     bool
)

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func existsDir(path string) (bool, error) {
	ex, err := exists(path)
	if !ex || err != nil {
		return ex, err
	}

	st, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return st.IsDir(), nil
}

func usage() {
	fmt.Printf("*** %v - %v ***\n\n", programName, programDescription)
	fmt.Println("Usage: ./makeaxe [OPTIONS] SOURCE [DESTINATION]")
	fmt.Println("OPTIONS")
	flag.VisitAll(func(f *flag.Flag) {
		if len(f.Name) < 2 {
			return
		}
		fmt.Printf("\t--%v, -%v\t%v\n", f.Name, f.Name[:1], f.Usage)
	})
	fmt.Println("ARGUMENTS")
	fmt.Println("\tSOURCE\t\tMandatory, the path of the unpackaged base directory. " +
		"\n\t\t\tWhen building a single axe, this should be the parent of the directory that contains a metadata.json file. " +
		"\n\t\t\tIf building all resolvers (--all) this should be the parent directory of all the resolvers.")

	fmt.Println("\tDESTINATION\tOptional, the path of the directory where newly built bundles (axes) should be placed. " +
		"\n\t\t\tIf unset, it is the same as the source directory.")
}

func version() {
	fmt.Printf("%v, version %v\n", programName, programVersion)
}

func bail(message string) {
	fmt.Println(message)
	fmt.Println("See ./makeaxe --help for usage information.")
	os.Exit(2)
}

func init() {
	const (
		flagAllUsage     = "build all the resolvers in the SOURCE path's subdirectories"
		flagReleaseUsage = "skip trying to add the git revision hash to a bundle"
		flagForceUsage   = "build a bundle and overwrite even if the destination directory already contains a bundle of the same name and version"
		flagHelpUsage    = "this help message"
		flagVersionUsage = "show version information"
	)
	flag.BoolVar(&all, "all", false, flagAllUsage)
	flag.BoolVar(&all, "a", false, flagAllUsage+" (shorthand)")
	flag.BoolVar(&release, "release", false, flagReleaseUsage)
	flag.BoolVar(&release, "r", false, flagReleaseUsage+" (shorthand)")
	flag.BoolVar(&force, "force", false, flagForceUsage)
	flag.BoolVar(&force, "f", false, flagForceUsage)
	flag.BoolVar(&help, "help", false, flagHelpUsage)
	flag.BoolVar(&help, "h", false, flagHelpUsage)
	flag.BoolVar(&ver, "version", false, flagVersionUsage)
	flag.BoolVar(&ver, "v", false, flagVersionUsage)

}

func main() {
	flag.Parse()

	if help {
		usage()
		return
	}

	if ver {
		version()
		return
	}

	if len(flag.Args()) == 0 {
		bail("Error: a source directory must be specified.")
	}

	if len(flag.Args()) > 2 {
		bail("Error: too many arguments.")
	}

	inputPath, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		bail("Error: bad source directory path.")
	}
	if ex, err := existsDir(inputPath); !ex || err != nil {
		bail("Error: bad source directory path.")
	}

	var outputPath string
	if len(flag.Args()) == 1 {
		outputPath = inputPath
	} else { //len is 2
		outputPath, err = filepath.Abs(flag.Arg(1))
		if err != nil {
			bail("Error: bad destination directory path.")
		}
		if ex, err := existsDir(outputPath); !ex || err != nil {
			bail("Error: bad destination directory path.")
		}
	}

	bundle.Make()

}
