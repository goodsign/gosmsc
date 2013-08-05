package rpcservice

import (
	log "github.com/cihub/seelog"
)

var logger log.LoggerInterface

func init() {
	DisableLog()
}

// DisableLog disables all library log output
func DisableLog() {
	logger = log.Disabled
}

// UseLogger uses a specified seelogger.LoggerInterface to output library logger.
// Use this func if you are using Seelog logging system in your app.
func UseLogger(newLogger log.LoggerInterface) {
	logger = newLogger
}

// Call this before app shutdown
func FlushLog() {
	logger.Flush()
}
