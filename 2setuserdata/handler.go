package main

import (
	"encoding/json"
	"thin-peak/httpservice"
	"thin-peak/logs/logger"

	"github.com/big-larry/mgo"
	"github.com/big-larry/suckhttp"
	"github.com/tarantool/go-tarantool"
	"go.mongodb.org/mongo-driver/bson"
)

type user struct {
	Id       string `bson:"_id"`
	Mail     string `bson:"mail"`
	Name     string `bson:"name"`
	Surname  string `bson:"surname"`
	Otch     string `bson:"otch"`
	Position string `bson:"position"`
	MetaId   string `bson:"metaid"`
	//Kaf      string `bson:"kaf"`
	//Fac      string `bson:"fac"`
}

type SetUserData struct {
	mgoSession      *mgo.Session
	mgoColl         *mgo.Collection
	trntlConn       *tarantool.Connection
	trntlTable      string
	trntlCodesTable string
	tokenGenerator  *httpservice.InnerService
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

	if len(r.Body) == 0 {
		return suckhttp.NewResponse(400, "Bad Request"), nil
	}
	userDataMarshalled := r.Body
	userData := &user{}
	err := json.Unmarshal(userDataMarshalled, userData)
	if err != nil {
		return nil, err
	}

	change := mgo.Change{
		Update:    bson.M{"$addToSet": bson.M{"metas": &meta{Type: 1, Id: fid}}},
		Upsert:    false
		ReturnNew: true,
		Remove:    false,
	}
	var foo interface{}

	_, err = conf.mgoColl.Find(query).Apply(change, &foo)
	if err != nil {
		if err == mgo.ErrNotFound {
			return suckhttp.NewResponse(400, "Bad request"), nil
		}
		return nil, err
	}

	// mongo insert
	insertData := &User{Id: userMailHash, Mail: userMail, Surname: userF, Name: userI, Otch: userO, Position: userPosition, MetaId: trntlRes[0].MetaId}

	err = conf.mgoColl.Insert(insertData)
	if err != nil {
		_, errr := conf.trntlConn.Delete(conf.trntlTable, "primary", []interface{}{userMailHash})
		if errr != nil {
			l.Error("Mongo insert", err)
			return nil, errr
		}
		return nil, err
	}

	return
}
