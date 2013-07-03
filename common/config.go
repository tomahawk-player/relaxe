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

package common

import (
	"encoding/json"
	"fmt"
	"github.com/teo/jsonmin"
	"io/ioutil"
)

type RelaxeConfig struct {
	Database struct {
		ConnectionString string
	}
	CacheDirectory string
}

func LoadConfig(path string) (*RelaxeConfig, error) {
	configFileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error: cannot read Relaxe configuration file.")
	}

	configFileBytes, _ = jsonmin.Minify(configFileBytes, false)
	var config RelaxeConfig
	err = json.Unmarshal(configFileBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("Error: bad configuration file format.")
	}
	return &config, nil
}
