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
	"encoding/json"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"github.com/teo/relaxe/common"
	"github.com/teo/relaxe/common/util"
	"github.com/teo/relaxe/makeaxe/bundle"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"path"
	"strings"
)

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

	log.Println("Connected to Relaxe MongoDB instance, collections: " + c.FullName)

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
		err = c.Insert(b.Metadata)
		if err != nil {
			log.Println(err.Error())
		}

		built = append(built, "UUID:"+axeUuid+"\t"+b.Metadata.PluginName+"-"+b.Metadata.Version)
	}

	if ex, err := util.ExistsFile(path.Join(outputPath, "index.html")); !ex && err == nil {
		indexText := fmt.Sprintf("<html><head><title>Relaxe server</title></head>" +
			"<body>Relaxe cache directory. Move along, nothing to see here.</body></html>")
		err := ioutil.WriteFile(path.Join(outputPath, "index.html"), []byte(indexText), 0644)
		if err != nil {
			log.Printf("Warning: could not write Relaxe index file.\n")
		}
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
			log.Printf("\tStatus: %v", err)
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
