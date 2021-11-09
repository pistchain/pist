package testlog

import "git.taiyue.io/pist/go-pist/log"

var msg string = "P2P"

func AddLog(ctx ...interface{}) {
	log.Info(msg, ctx...)
}
