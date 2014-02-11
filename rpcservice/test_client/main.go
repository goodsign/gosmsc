package main

import (
	log "github.com/cihub/seelog"
	. "github.com/goodsign/gosmsc/rpcservice/client"
	"github.com/goodsign/goutils/jsonrpc"
	"os"
	"time"
)

const (
	SeelogCfg = "seelog.xml"
)

func loadLogger() {
	logger, err := log.LoggerFromConfigAsFile(SeelogCfg)
	if err != nil {
		panic(err)
	}
	jsonrpc.UseLogger(logger)
	log.ReplaceLogger(logger)
}

func main() {
	loadLogger()
	defer log.Flush()

	c, err := NewSmscRpcServiceClient("http://localhost:5678/rpc", 20, time.Minute)
	if err != nil {
		log.Critical(err)
		log.Flush()
		os.Exit(-1)
	}
	id, err := c.Send("+79211234567", "test", true)
	if err != nil {
		log.Critical(err)
		log.Flush()
		os.Exit(-1)
	}
	log.Infof("Id: '%d'\n", id)

	for i := 0; i < 100; i++ {
		time.Sleep(time.Second)
		st, err := c.GetActualStatus(id)
		if err != nil {
			log.Error(err)
		} else {
			log.Infof("Status: '%d'", st.StatusCode)
		}
	}

}
