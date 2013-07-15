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

import ()

type Axe_v1 struct { // deprecated
	Name            string `json:"name"`
	Author          string `json:"author"`
	BundleVersion   string `json:"bundleVersion"`
	Description     string `json:"description"`
	Email           string `json:"email"`
	Platform        string `json:"platform"`
	PluginName      string `json:"pluginName"`
	Revision        string `json:"revision,omitempty"`
	Timestamp       string `json:"timestamp"`
	TomahawkVersion string `json:"tomahawkVersion"`
	Type            string `json:"type"`
	Version         string `json:"version"`
	Website         string `json:"website"`
	Manifest        struct {
		Icon      string   `json:"icon"`
		Main      string   `json:"main"`
		Scripts   []string `json:"scripts"`
		Resources []string `json:"resources"`
	} `json:"manifest"`
}

type Axe_v2 struct {
	PluginName string `json:"pluginName"`
	Name       string `json:"name"`
	Author     string `json:"author,omitempty"` //deprecated
	Email      string `json:"email,omitempty"`  //deprecated
	Authors    []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"authors"`
	License           string    `json:"license"` //Allowed values: GPL3, BSD, MIT, X11, ...
	CustomLicenseText string    `json:"customLicenseText,omitempty" bson:",omitempty"`
	BundleVersion     string    `json:"bundleVersion"`
	Description       string    `json:"description"`
	Platform          string    `json:"platform"`
	Revision          string    `json:"revision,omitempty" bson:",omitempty"`
	Timestamp         *int64    `json:"timestamp,omitempty"` //nullable
	ApiVersion        string    `json:"apiVersion"`
	Version           string    `json:"version"`
	Website           string    `json:"website"`
	Type              string    `json:"type"` //Allowed values: resolver/javascript, resolver/binary
	Manifest          *struct { //ptr to make it nullable
		Icon      string   `json:"icon"`
		Main      string   `json:"main"`
		Scripts   []string `json:"scripts"`
		Resources []string `json:"resources"`
	} `json:"manifest,omitempty"`
	Features        []string `json:"features,omitempty" bson:",omitempty"`        //only if type == resolver/javascript
	BinarySignature string   `json:"binarySignature,omitempty" bson:",omitempty"` //only if type == resolver/binary

	// Only used on Relaxe, do *not* set in source metadata.json
	AxeId     string `json:"axeId,omitempty"`
	Downloads *int64 `json:"downloads,omitempty"`
}

func Axe_v2check(axe *Axe_v2) bool {
	if axe.PluginName == "" ||
		axe.Name == "" ||
		axe.Version == "" ||
		axe.Description == "" ||
		axe.Type == "" {
		return false
	}

	if axe.Type != "resolver/javascript" &&
		axe.Type != "resolver/binary" {
		return false
	}

	if axe.Type == "resolver/javascript" &&
		(axe.Manifest.Main == "" ||
			axe.Manifest.Icon == "") {
		return false
	}

	return true
}
