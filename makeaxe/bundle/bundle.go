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

package bundle

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/teo/relaxe/makeaxe/util"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const (
	bundleVersion = "1"
)

func Package(inputPath string, outputPath string, release bool) error {
	metadataRelPath := "content/metadata.json"
	metadataPath := path.Join(inputPath, metadataRelPath)

	ex, err := util.ExistsFile(metadataPath)
	if err != nil {
		return err
	}
	if !ex {
		return fmt.Errorf("Cannot find metadata file in %v. Make sure %v exists and is readable.",
			inputPath, metadataRelPath)
	}

	metadataBytes, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return err
	}

	metadata := make(map[string]interface{})
	json.Unmarshal(metadataBytes, &metadata)
	fmt.Println(metadata)

	if metadata["pluginName"] != nil &&
		metadata["name"] != nil &&
		metadata["version"] != nil &&
		metadata["description"] != nil &&
		metadata["type"] != nil &&
		metadata["manifest"] != nil &&
		metadata["manifest"].(map[string]interface{})["main"] != nil &&
		metadata["manifest"].(map[string]interface{})["icon"] != nil {
		fmt.Printf("* Metadata for bundle %v-%v looks ok.\n", metadata["pluginName"], metadata["version"])
	} else {
		return fmt.Errorf("Bad metadata file in %v.", metadataPath)
	}
	pluginName := metadata["pluginName"].(string)
	version := metadata["version"].(string)

	// Let's add some stuff to the metadata file, this is information that's much
	// easier to fill in automatically now than manually whenever.
	//   * Timestamp of right now i.e. packaging time.
	//   * Git revision because it makes sense, especially during development.
	//   * Bundle format version, which might never be used but we add it just in
	//     case we ever need to distinguish one bundle format from another.
	//	 * Strip comments.
	metadata["timestamp"] = time.Now().Unix()
	metadata["bundleVersion"] = bundleVersion
	if !release {
		gitCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
		gitCmd.Dir = inputPath
		revision, err := gitCmd.Output()
		if err == nil { //we are in a git repo
			metadata["revision"] = strings.TrimSpace(string(revision))
		} else {
			fmt.Printf("Warning: cannot get revision hash for %v-%v.", pluginName, version)
		}
	}

	metadataToWrite, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	// Let's do some zipping according to the manifest.
	filesToZip := []string{}
	m := metadata["manifest"].(map[string]interface{})
	filesToZip = append(filesToZip, path.Join("content", m["main"].(string)))
	if m["scripts"] != nil {
		for _, s := range m["scripts"].([]interface{}) {
			filesToZip = append(filesToZip, path.Join("content", s.(string)))
		}
	}
	filesToZip = append(filesToZip, path.Join("content", m["icon"].(string)))
	if m["resources"] != nil {
		for _, s := range m["resources"].([]interface{}) {
			filesToZip = append(filesToZip, path.Join("content", s.(string)))
		}
	}

	outputFileName := pluginName + "-" + version + ".axe"
	outputFilePath := path.Join(outputPath, outputFileName)

	ex, err = util.ExistsFile(outputFilePath)
	if ex || err != nil {
		if err := os.Remove(outputFilePath); err != nil {
			return err
		}
	}

	f, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	z := zip.NewWriter(f)
	defer z.Close()
	for _, fileName := range filesToZip {
		currentFile, err := z.Create(fileName)
		if err != nil {
			return err
		}
		body, err := ioutil.ReadFile(path.Join(inputPath, fileName))
		if err != nil {
			return err
		}
		_, err = currentFile.Write(body)
		if err != nil {
			return err
		}
	}
	currentFile, err := z.Create(metadataRelPath)
	if err != nil {
		return err
	}
	_, err = currentFile.Write(metadataToWrite)
	if err != nil {
		return err
	}

	sumFile, err := util.Md5sum(outputFilePath)
	if err != nil {
		fmt.Printf("Warning: could not create MD5 hash file for %v.", outputFileName)
	}
	sumFile += "\t" + outputFileName
	sumFilePath := path.Join(outputPath, pluginName+"-"+version+".md5")
	err = ioutil.WriteFile(sumFilePath, []byte(sumFile), 0644)

	fmt.Printf("* Created axe in %v.", outputFilePath)

	return nil
}
