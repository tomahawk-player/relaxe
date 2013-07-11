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
	"github.com/coocood/jas"
	"github.com/teo/relaxe/common"
	"github.com/teo/relaxe/common/util"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
)

const (
	programName        = "Relaxe"
	programDescription = "the Tomahawk axe server"
	programVersion     = "0.1"
)

var (
	help bool
)

func usage() {
	fmt.Printf("*** %v %v - %v ***\n\n", programName, programVersion, programDescription)
	fmt.Println("Usage: ./relaxe [OPTIONS] [CONFIG]")
	fmt.Println("OPTIONS")
	flag.VisitAll(func(f *flag.Flag) {
		if len(f.Name) < 2 || f.Name[:4] == "test" {
			return
		}
		fmt.Printf("\t%v\n", f.Usage)
	})

	fmt.Println("ARGUMENTS")
	fmt.Println("\tCONFIG\t\tThe path of the Relaxe configuration file, defaults to \"./relaxe.json\".")
}

func init() {
	const (
		flagHelpUsage = "--help, -h\tthis help message"
	)
	flag.BoolVar(&help, "help", false, flagHelpUsage)
	flag.BoolVar(&help, "h", false, flagHelpUsage)

	flag.Usage = usage
}

func sigintCatcher() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	os.Exit(0)
}

func die(message string) {
	fmt.Println(message)
	fmt.Println("See ./relaxe --help for usage information.")
	os.Exit(2)
}

func main() {
	flag.Parse()

	if help {
		usage()
		return
	}

	configFilePath := ""
	var err error

	switch len(flag.Args()) {
	case 0:
		configFilePath, err = os.Getwd()
		if err != nil {
			panic(err)
		}
		configFilePath = path.Join(configFilePath, "relaxe.json")
	case 1:
		configFilePath, err = filepath.Abs(flag.Arg(0))
		if err != nil {
			panic(err)
		}
	default:
		die("Error: too many arguments.")
	}

	if ex, err := util.ExistsFile(configFilePath); !ex || err != nil {
		die("Bad Relaxe configuration file path: " + configFilePath)
	}

	go sigintCatcher()

	config, err := common.LoadConfig(configFilePath)
	if err != nil {
		fmt.Println(err.Error())
		die("Cannot load config file.")
	}

	// Jas router for the API
	axes, err := NewAxes(config)
	if err != nil {
		die("Error: cannot start Relaxe server. Reason: " + err.Error())
	}

	router := jas.NewRouter(axes)
	router.RequestErrorLogger = router.InternalErrorLogger
	router.BasePath = "/v1/"

	// FileServer for the axes cache
	fileserver := http.StripPrefix(config.Server.CachePath,
		http.FileServer(http.Dir(config.CacheDirectory)))

	hostString := fmt.Sprintf("%v:%v", config.Server.Host, config.Server.Port)

	helloMessage := fmt.Sprintf("Starting Relaxe server on %v.\n", hostString) +
		fmt.Sprintf("Relaxe serving paths:\n%v\n", router.HandledPaths(true)) +
		fmt.Sprintf("Axes cache:\t%v\n", config.Server.CachePath)

	log.Print(helloMessage)

	http.Handle(router.BasePath, router)
	http.Handle(config.Server.CachePath, fileserver)
	err = http.ListenAndServe(hostString, nil)
	if err != nil {
		panic(err)
	}

}
