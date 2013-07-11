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
	"github.com/coocood/jas"
	"github.com/teo/relaxe/common"
	"github.com/teo/relaxe/common/util"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"path"
)

type Axes struct {
	config *common.RelaxeConfig
	c      *mgo.Collection
}

func NewAxes(config *common.RelaxeConfig) (*Axes, error) {
	this := new(Axes)
	this.config = config

	// hook up to db
	session, err := mgo.Dial(config.Database.ConnectionString)

	this.c = session.DB("relaxe").C("axes")

	return this, err
}

func (*Axes) Gap() string {
	return ":tomahawkVersion/:platform/:name"
}

func newestAxe(axes []common.Axe_v2) *common.Axe_v2 {
	currentVersion := "0"
	newestAxe := 0

	for i, _ := range axes {
		if util.VersionCompare(axes[i].Version, currentVersion) > 0 {
			currentVersion = axes[i].Version
			newestAxe = i
		}
	}

	return &axes[newestAxe]
}

func (this *Axes) Get(ctx *jas.Context) { // `GET /axes`
	tomahawkVersion := ctx.GapSegment(":tomahawkVersion")
	platform := ctx.GapSegment(":platform")
	name := ctx.GapSegment(":name")

	response := []common.Axe_v2{}
	var err error

	if name == "" {
		err = this.c.Find(bson.M{"platform": bson.M{"$in": []string{"", "any", platform}}}).All(&response)
	} else { //name not empty
		err = this.c.Find(bson.M{"pluginname": name,
			"platform": bson.M{"$in": []string{"", "any", platform}}}).All(&response)
	}

	if err != nil {
		log.Println(err.Error())
	}

	// apply version filters
	entries := map[string][]common.Axe_v2{}
	for _, axe := range response {
		if tomahawkVersion == "" || axe.TomahawkVersion == "" ||
			util.VersionCompare(tomahawkVersion, axe.TomahawkVersion) >= 0 {
			if entries[axe.PluginName] == nil {
				entries[axe.PluginName] = []common.Axe_v2{}
			}
			entries[axe.PluginName] = append(entries[axe.PluginName], axe)
		}
	}

	response = []common.Axe_v2{}
	for _, axes := range entries {
		response = append(response, *newestAxe(axes))
	}

	if name == "" {
		for i, _ := range response {
			response[i].Timestamp = nil
			response[i].Manifest = nil
			response[i].AxeId = ""
			response[i].Features = []string{}
			//don't ship legacy-formatted info
			response[i].Author = ""
			response[i].Email = ""
		}

		ctx.Data = response
	} else {
		if len(response) != 1 {
			log.Println("Error: bad entry count for pluginName " + name)
			return
		}
		realResponse := map[string]string{}
		realResponse["pluginName"] = response[0].PluginName
		realResponse["version"] = response[0].Version
		axeFilename := response[0].PluginName + "-" + response[0].AxeId + ".axe"
		realResponse["contentPath"] = path.Join(config.Server.CachePath, axeFilename)
		ctx.Data = realResponse
	}

	if ctx.Error != nil {
		log.Println(ctx.Error)
	}
}

// `GET /axes/:version/:platform/` 			==> []Axe_v2 trimmed
// `GET /axes/:version/:platform/:name` 	==> { pluginName, version, contentPath }
