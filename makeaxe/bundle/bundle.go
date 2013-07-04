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

package bundle

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/teo/relaxe/common"
	"github.com/teo/relaxe/makeaxe/util"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const (
	bundleVersion   = "2"
	metadataRelPath = "content/metadata.json"
)

type Bundle struct {
	Metadata     *common.Axe_v2
	InputDirPath string
}

func LoadBundle(inputDirPath string) (*Bundle, error) {
	b := new(Bundle)
	b.InputDirPath = inputDirPath

	metadataPath := path.Join(inputDirPath, metadataRelPath)

	ex, err := util.ExistsFile(metadataPath)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, fmt.Errorf("Cannot find metadata file in %v. Make sure %v exists and is readable.",
			inputDirPath, metadataRelPath)
	}

	metadataBytes, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	metadata := new(common.Axe_v2)
	err = json.Unmarshal(metadataBytes, metadata)

	if err != nil || !common.Axe_v2check(metadata) {
		return nil, fmt.Errorf("Bad metadata file in %v.", metadataPath)
	}

	// mangle a bit for backwards compatibility with v1 metadata.json
	if metadata.Author != "" || metadata.Email != "" {
		fmt.Printf("Warning: author and email fields for %v are deprecated.\n", metadata.PluginName)
		if len(metadata.Authors) == 0 {
			metadata.Authors = append(metadata.Authors, struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}{metadata.Author, metadata.Email})
		}
	}

	if metadata.License == "" {
		fmt.Printf("Warning: license field is empty for %v.\n", metadata.PluginName)
	}

	// Bundle version to distinguish one bundle format from another.
	metadata.BundleVersion = bundleVersion

	b.Metadata = metadata
	return b, nil
}

func (this *Bundle) CreatePackage(outputDirPath string, release bool, force bool) (string, error) {
	metadata := this.Metadata
	pluginName := metadata.PluginName
	version := metadata.Version

	var (
		outputFileName string
		sumFileName    string
	)

	if metadata.AxeId != "" {
		outputFileName = pluginName + "-" + metadata.AxeId + ".axe"
		sumFileName = pluginName + "-" + metadata.AxeId + ".md5"
	} else {
		outputFileName = pluginName + "-" + version + ".axe"
		sumFileName = pluginName + "-" + version + ".md5"
	}
	outputFilePath := path.Join(outputDirPath, outputFileName)

	ex, err := util.ExistsFile(outputFilePath)
	if !force && (ex || err != nil) { //if we don't force, and the target either exists or we're not sure
		log.Printf("* %v already exists, skipping.\n", outputFileName)
		return outputFilePath, fmt.Errorf("Axe file %v already exists, skipping.", outputFileName)
	}

	// Let's add some stuff to the metadata file, this is information that's much
	// easier to fill in automatically now than manually whenever.
	//   * Timestamp of right now i.e. packaging time.
	//   * Git revision because it makes sense, especially during development.
	metadata.Timestamp = time.Now().Unix()
	if !release {
		gitCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
		gitCmd.Dir = this.InputDirPath
		revision, err := gitCmd.Output()
		if err == nil { //we are in a git repo
			metadata.Revision = strings.TrimSpace(string(revision))
		} else {
			log.Printf("Warning: cannot get revision hash for %v-%v.\n", pluginName, version)
		}
	}

	metadataToWrite, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", err
	}

	// Let's do some zipping according to the manifest.
	filesToZip := []string{}
	m := metadata.Manifest
	filesToZip = append(filesToZip, path.Join("content", m.Main))
	if m.Scripts != nil {
		for _, s := range m.Scripts {
			filesToZip = append(filesToZip, path.Join("content", s))
		}
	}
	filesToZip = append(filesToZip, path.Join("content", m.Icon))
	if m.Resources != nil {
		for _, s := range m.Resources {
			filesToZip = append(filesToZip, path.Join("content", s))
		}
	}

	ex, err = util.ExistsFile(outputFilePath)
	if ex || err != nil {
		if err := os.Remove(outputFilePath); err != nil {
			return "", err
		}
	}

	f, err := os.Create(outputFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	z := zip.NewWriter(f)
	defer z.Close()
	for _, fileName := range filesToZip {
		currentFile, err := z.Create(fileName)
		if err != nil {
			return "", err
		}
		body, err := ioutil.ReadFile(path.Join(this.InputDirPath, fileName))
		if err != nil {
			return "", err
		}
		_, err = currentFile.Write(body)
		if err != nil {
			return "", err
		}
	}
	currentFile, err := z.Create(metadataRelPath)
	if err != nil {
		return "", err
	}
	_, err = currentFile.Write(metadataToWrite)
	if err != nil {
		return "", err
	}

	sumValue, err := util.Md5sum(outputFilePath)
	if err != nil {
		log.Printf("Warning: could not create MD5 hash file for %v.\n", outputFileName)
	}
	sumValue += "\t" + outputFileName
	sumFilePath := path.Join(outputDirPath, sumFileName)
	err = ioutil.WriteFile(sumFilePath, []byte(sumValue), 0644)

	return outputFilePath, nil
}
