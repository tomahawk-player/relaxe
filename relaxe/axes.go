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
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
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

func (this *Axes) Get(ctx *jas.Context) { // `GET /axes`
	response := []common.Axe_v2{}
	err := this.c.Find(bson.M{}).All(&response)

	if err != nil {
		log.Println(err.Error())
	}

	for i, _ := range response {
		response[i].Timestamp = nil
		response[i].Manifest = nil
		//don't ship legacy-formatted info
		response[i].Author = ""
		response[i].Email = ""
	}

	ctx.Data = response
	log.Println(ctx.ResponseHeader.Get("content-type") + " body: " + ctx.Data.(string))
}
