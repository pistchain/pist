package help

import "git.taiyue.io/pist/go-pist/log"

func CheckAndPrintError(err error) {
	if err != nil {
		log.Debug("CheckAndPrintError", "error", err.Error())
	}
}
