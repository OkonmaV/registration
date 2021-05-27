package main

import (
	"encoding/json"
	"strings"
	"thin-peak/logs/logger"

	"github.com/big-larry/mgo"
	"github.com/big-larry/suckhttp"
	"go.mongodb.org/mongo-driver/bson"
)

type user struct {
	Id       string `bson:"_id" json:"_id"`
	Mail     string `bson:"mail" json:"mail,omitempty"`
	Name     string `bson:"name" json:"name,omitempty"`
	Surname  string `bson:"surname" json:"surname,omitempty"`
	Otch     string `bson:"otch" json:"otch,omitempty"`
	Position string `bson:"position" json:"position,omitempty"`
	MetaId   string `bson:"metaid" json:"metaid,omitempty"`
}

type SetUserData struct {
	mgoSession *mgo.Session
	mgoColl    *mgo.Collection
}

func NewSetUserData(mgodb string, mgoAddr string, mgoColl string) (*SetUserData, error) {

	mgoSession, err := mgo.Dial(mgoAddr)
	if err != nil {
		return nil, err
	}
	logger.Info("Mongo", "Connected!")

	return &SetUserData{mgoSession: mgoSession, mgoColl: mgoSession.DB(mgodb).C(mgoColl)}, nil
}

func (c *SetUserData) Close() error {
	c.mgoSession.Close()
	return nil
}

func (conf *SetUserData) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	if !strings.Contains(r.GetHeader(suckhttp.Content_Type), "application/json") {
		l.Debug("Content-type", "Wrong content-type at POST")
		return suckhttp.NewResponse(400, "Bad request"), nil
	}
	if len(r.Body) == 0 {
		return suckhttp.NewResponse(400, "Bad Request"), nil
	}
	var upsertData map[string]interface{}
	err := json.Unmarshal(r.Body, &upsertData)
	if err != nil {
		return suckhttp.NewResponse(400, "Bad request"), err
	}
	if upsertData["_id"] == nil {
		l.Debug("Request json", "\"_id\" field is nil")
		return suckhttp.NewResponse(400, "Bad request"), nil
	}

	update := bson.M{"$set": &upsertData, "$currentDate": bson.M{"lastmodified": true}}

	_, err = conf.mgoColl.Upsert(&bson.M{"_id": upsertData["_id"], "deleted": bson.M{"$exists": false}}, update)
	if err != nil {
		return nil, err
	}

	return suckhttp.NewResponse(200, "OK"), nil

}
