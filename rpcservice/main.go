package main

import (
	"fmt"
	"os"
	"net/http"
	"github.com/goodsign/rpc"
	gjson "github.com/goodsign/rpc/json"
	"flag"
	"io/ioutil"
	"encoding/json"
	"github.com/goodsign/gosmsc"
	log "github.com/cihub/seelog"
)

const (
	ErrorCodeInvalidConfig = -1
	ErrorCodeInvalidArgs = -2
	ErrorCodeInternalInitError = -3

	SeelogCfg = "seelog.xml"
)

var (
	rpcPath = flag.String("rpcpath", "rpc", "Rpc service path (http.Handle parameter)")
	port    = flag.String("p", "5678", "Port")
	cfgPath = flag.String("cfg", "smsc.default.json", "Path to service configuration file")
)

func loadLogger() {
	logger, err := log.LoggerFromConfigAsFile(SeelogCfg)

	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)
}

func unmarshalConfig(configFileName string) (conf *gosmsc.Sender, err error) {
	log.Infof("loading config from %s", configFileName)

	bytes, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return nil, err
	}

	opts := new(gosmsc.SenderOptions)
	log.Debug("Unmarshalling config")
	err = json.Unmarshal(bytes, opts)
	if err != nil {
		return nil, fmt.Errorf("Cannot unmarshal: '%s'", err)
	}
	conf, err = gosmsc.NewSender(opts)
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
	s.RegisterService(&SMSService{sender}, "")

	http.Handle("/" + *rpcPath, s)

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

//Service Definition
type SMSService struct {
	sender *gosmsc.Sender
}

type Send_Args struct {
	Phone string
	Text  string
}

type Send_Reply struct {
	Message string
}

func (h *SMSService) Send(r *http.Request, msg *Send_Args, reply *Send_Reply) error {
	log.Trace("")

	err := h.sender.Send(msg.Phone, msg.Text)
	if err != nil {
		reply.Message = err.Error()
		return err
	}
	reply.Message = "OK"
	return nil
}
