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

package common

import (
	"encoding/json"
	"fmt"
	"github.com/teo/jsonmin"
	"io/ioutil"
)

type RelaxeConfig struct {
	CacheDirectory string `json:"cacheDirectory"`
	Database       struct {
		ConnectionString string `json:"connectionString"`
	} `json:"database"`
	KvStore struct {
		ConnectionString string `json:"connectionString"`
	} `json:"kvStore"`
	Server struct {
		Host      string `json:"host"`
		Port      uint16 `json:"port"`
		CachePath string `json:"cachePath"`
	} `json:"server"`
}

func LoadConfig(path string) (*RelaxeConfig, error) {
	configFileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error: cannot read Relaxe configuration file: " + path)
	}

	configFileBytes, _ = jsonmin.Minify(configFileBytes, false)
	var config RelaxeConfig
	err = json.Unmarshal(configFileBytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
