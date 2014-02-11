package main

import (
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/goodsign/gosmsc"
	"github.com/goodsign/gosmsc/rpcservice"
	"github.com/goodsign/goutils/mgo"
	"github.com/goodsign/rpc"
	gjson "github.com/goodsign/rpc/json"
	"io/ioutil"
	lmgo "labix.org/v2/mgo"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	ErrorCodeInvalidConfig     = -1
	ErrorCodeInvalidArgs       = -2
	ErrorCodeInternalInitError = -3
	ConnectTimeout             = 5 * time.Minute

	SeelogCfg = "seelog.xml"
)

var (
	rpcPath        = flag.String("rpcpath", "rpc", "Rpc service path (http.Handle parameter)")
	port           = flag.String("p", "5678", "Port")
	cfgPath        = flag.String("cfg", "smsc.default.json", "Path to service configuration file")
	mongoPath      = flag.String("dbpath", "localhost", "Mongo path")
	mongoDb        = flag.String("mongodb", "gastody_sms_service", "Mongo DB")
	updateInterval = flag.String("interval", "60000", "Update interval in milliseconds")
)

func loadLogger() {
	logger, err := log.LoggerFromConfigAsFile(SeelogCfg)

	if err != nil {
		panic(err)
	}
	rpcservice.UseLogger(logger)
	gosmsc.UseLogger(logger)
	log.ReplaceLogger(logger)
}

func unmarshalConfig(configFileName string) (conf *gosmsc.SenderCheckerImpl, err error) {
	log.Infof("loading config from %s", configFileName)

	bytes, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return nil, err
	}

	opts := new(gosmsc.SmscClientOptions)
	log.Debug("Unmarshalling config")
	err = json.Unmarshal(bytes, opts)
	if err != nil {
		return nil, fmt.Errorf("Cannot unmarshal: '%s'", err)
	}
	dinfo := &lmgo.DialInfo{
		[]string{*mongoPath},
		true,
		ConnectTimeout,
		false,
		*mongoDb,
		"",
		"",
		nil,
		nil,
	}
	hlp, err := mgo.Dial(dinfo, &mgo.DbHelperInitOptions{&lmgo.Safe{}})
	if err != nil {
		return nil, err
	}
	str, err := gosmsc.NewMessageStatusMgoStorage(hlp)
	if err != nil {
		return nil, err
	}
	upint, err := strconv.ParseInt(*updateInterval, 10, 32)
	if err != nil {
		return nil, err
	}
	conf, err = gosmsc.NewSenderCheckerImpl(opts, str, time.Millisecond*time.Duration(upint))
	if err != nil {
		return nil, fmt.Errorf("Invalid config: '%s'", err)
	}

	return
}

func fail(code int, msg string) {
	log.Critical(msg)
	log.Flush()
	os.Exit(code)
}

func main() {
	flag.Parse()

	loadLogger()
	defer log.Flush()

	if len(*rpcPath) == 0 {
		fail(ErrorCodeInvalidArgs, "Please specify rpc path")
	}
	if len(*port) == 0 {
		fail(ErrorCodeInvalidArgs, "Please specify port")
	}
	if len(*cfgPath) == 0 {
		fail(ErrorCodeInvalidArgs, "Please specify config file path")
	}

	sender, err := unmarshalConfig(*cfgPath)
	if err != nil {
		fail(ErrorCodeInvalidConfig, fmt.Sprintf("Sender init failed. '%s'", err))
	}

	s := rpc.NewServer()
	s.RegisterCodec(gjson.NewCodec(), "application/json")
	if err != nil {
		fail(ErrorCodeInternalInitError, err.Error())
	}

	serv, err := rpcservice.NewSMSService(sender)
	s.RegisterService(serv, "")

	http.Handle("/"+*rpcPath, s)

	ml, err := s.ListMethods("SMSService")
	if err != nil {
		fail(ErrorCodeInternalInitError, err.Error())
	}
	str := fmt.Sprintf("\nStarting service '/%s' on port ':%s'. \nMethods:\n",
		*rpcPath, *port)
	for _, m := range ml {
		str = str + "    " + m + "\n"
	}
	log.Info(str)
	log.Flush()

	err = http.ListenAndServe(":"+*port, nil)
	if err != nil {
		fail(ErrorCodeInternalInitError, err.Error())
	}
}
