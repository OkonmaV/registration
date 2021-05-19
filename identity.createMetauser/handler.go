package main

import (
	"encoding/json"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/mgo"
	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/rs/xid"
	"github.com/tarantool/go-tarantool"
	"go.mongodb.org/mongo-driver/bson"
)

type CreateMetauser struct {
	trntlConn  *tarantool.Connection
	trntlTable string
	mgoSession *mgo.Session
	mgoColl    *mgo.Collection
}

type metauser struct {
	MetaId  string `json:"metaid"`
	RegCode string `json:"regcode"`
}

func NewCreateMetauser(trntlAddr string, trntlTable string, mgoAddr string, mgoColl string) (*CreateMetauser, error) {
	trntlConnection, err := tarantool.Connect(trntlAddr, tarantool.Opts{
		// User: ,
		// Pass: ,
		Timeout:       500 * time.Millisecond,
		Reconnect:     1 * time.Second,
		MaxReconnects: 4,
	})
	if err != nil {
		logger.Error("Tarantool", err)
		return nil, err
	}
	logger.Info("Tarantool", "Connected!")

	mgoSession, err := mgo.Dial(mgoAddr)
	if err != nil {
		logger.Error("Mongo conn", err)
		return nil, err
	}
	logger.Info("Mongo", "Connected!")
	mgoCollection := mgoSession.DB("main").C(mgoColl)

	return &CreateMetauser{trntlConn: trntlConnection, trntlTable: trntlTable, mgoSession: mgoSession, mgoColl: mgoCollection}, nil
}

func (handler *CreateMetauser) Close() error {
	handler.mgoSession.Close()
	return handler.trntlConn.Close()
}

func getRandId() string {
	return xid.New().String()
}

func (conf *CreateMetauser) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	// TODO: AUTH

	var metaSurname, metaName, contextFolderId string
	switch r.GetMethod() {
	case suckhttp.GET:
		metaSurname = r.Uri.Query().Get("surname")
		metaName = r.Uri.Query().Get("name")
		contextFolderId = r.Uri.Query().Get("contextid")
		if metaName == "" || metaSurname == "" || contextFolderId == "" {
			return suckhttp.NewResponse(400, "Bad Request"), nil
		}
	case suckhttp.POST:
		if !strings.Contains(r.GetHeader(suckhttp.Content_Type), "application/x-www-form-urlencoded") {
			return suckhttp.NewResponse(400, "Bad request"), nil
		}
		formValues, err := url.ParseQuery(string(r.Body))
		if err != nil {
			return suckhttp.NewResponse(400, "Bad Request"), err
		}
		metaSurname = formValues.Get("surname")
		metaName = formValues.Get("name")
		contextFolderId = formValues.Get("context")
		if metaName == "" || metaSurname == "" || contextFolderId == "" {
			return suckhttp.NewResponse(400, "Bad Request"), nil
		}
	}
	//TODO
	userMetaId := "randmetaid"
	//

	//TODO: maybe delete when auth done
	query := &bson.M{"_id": contextFolderId, "deleted": bson.M{"$exists": false}, "metas.metatype": 1, "metas.metaid": userMetaId}

	var foo interface{}
	err := conf.mgoColl.Find(query).One(&foo)
	if err != nil {
		if err == mgo.ErrNotFound {
			return suckhttp.NewResponse(403, "Forbidden"), nil
		}
		return nil, err
	}
	//

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	var code int
	metaId := getRandId()
	for {
		code = int(rnd.Int31n(90000) + 10000)
		_, err := conf.trntlConn.Insert(conf.trntlTable, []interface{}{code, metaId, metaSurname, metaName})
		if err != nil {
			if tarErr, ok := err.(*tarantool.Error); ok && tarErr.Code == tarantool.ErrTupleFound {
				continue
			} else {
				return nil, err
			}
		}
		break
	}

	resp := suckhttp.NewResponse(200, "OK")
	if r.GetMethod() == suckhttp.GET {
		var body []byte
		contentType := "text/plain"
		if strings.Contains(r.GetHeader(suckhttp.Accept), "application/json") {
			var err error
			body, err = json.Marshal(&metauser{MetaId: metaId, RegCode: strconv.Itoa(code)})
			if err != nil {
				_, errr := conf.trntlConn.Delete(conf.trntlTable, "primary", []interface{}{code})
				if errr != nil {
					l.Warning("Undeleted code from Tarantool:", strconv.Itoa(code))
					l.Error("Tarantool delete", errr)
					return nil, err
				}
				return nil, err
			}
			contentType = "application/json"
		} else {
			body = []byte(suckutils.ConcatThree(metaId, ";", strconv.Itoa(code)))
		}
		resp.AddHeader(suckhttp.Content_Type, suckutils.ConcatTwo(contentType, "; charset=utf8"))
		resp.SetBody(body)
	}

	return resp, nil
}
