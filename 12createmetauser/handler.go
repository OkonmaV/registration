package main

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"thin-peak/httpservice"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/mgo"
	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/rs/xid"
	"github.com/tarantool/go-tarantool"
)

type CreateMetauser struct {
	trntlConn      *tarantool.Connection
	trntlTable     string
	mgoSession     *mgo.Session
	mgoColl        *mgo.Collection
	codeGeneration *httpservice.InnerService
}

type metauser struct {
	MetaId  string `json:"metaid"`
	RegCode string `json:"regcode"`
	Surname string `json:"surname"`
	Name    string `json:"name"`
}

func NewCreateMetauser(trntlAddr string, trntlTable string, mgodb string, mgoAddr string, mgoColl string, codeGeneration *httpservice.InnerService) (*CreateMetauser, error) {
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
	mgoCollection := mgoSession.DB(mgodb).C(mgoColl)

	return &CreateMetauser{trntlConn: trntlConnection, trntlTable: trntlTable, mgoSession: mgoSession, mgoColl: mgoCollection, codeGeneration: codeGeneration}, nil
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

	if !strings.Contains(r.GetHeader(suckhttp.Content_Type), "application/x-www-form-urlencoded") {
		l.Debug("Content-type", "Wrong content-type at POST")
		return suckhttp.NewResponse(400, "Bad request"), nil
	}
	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return suckhttp.NewResponse(400, "Bad Request"), err
	}
	metaSurname := formValues.Get("surname")
	metaName := formValues.Get("name")
	//contextFolderId = formValues.Get("contextid")
	if metaName == "" || metaSurname == "" { //|| contextFolderId == "" {
		return suckhttp.NewResponse(400, "Bad Request"), nil
	}

	// //TODO: откуда мы берем метаид?
	// userMetaId := "randmetaid"
	// //

	// //TODO: maybe delete when auth done
	// query := &bson.M{"_id": contextFolderId, "deleted": bson.M{"$exists": false}, "metas.metatype": 1, "metas.metaid": userMetaId}

	metaId := getRandId()

	codeGenerationReq, err := conf.codeGeneration.CreateRequestFrom(suckhttp.POST, "", r)
	if err != nil {
		return nil, err
	}
	codeGenerationReq.Body = []byte(metaId)
	codeGenerationResp, err := conf.codeGeneration.Send(codeGenerationReq)
	if err != nil {
		return nil, err
	}

	// TODO: нужно ли эту херь вообще возвращать?
	resp := suckhttp.NewResponse(200, "OK")
	if r.GetMethod() == suckhttp.GET {
		var body []byte
		contentType := "text/plain"
		if strings.Contains(r.GetHeader(suckhttp.Accept), "application/json") {
			var err error
			body, err = json.Marshal(&metauser{MetaId: metaId, RegCode: strconv.Itoa(code), Surname: metaSurname, Name: metaName})
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
			body = []byte(metaId)
		}
		resp.AddHeader(suckhttp.Content_Type, suckutils.ConcatTwo(contentType, "; charset=utf8"))
		resp.SetBody(body)
	}

	return resp, nil
}
