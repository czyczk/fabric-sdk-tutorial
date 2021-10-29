package timingutils

import (
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	log "github.com/sirupsen/logrus"
)

func GetDeferrableTimingLogger(message string) func() {
	if !global.ShowTimingLogs {
		return func() {}
	}

	start := time.Now()
	return func() {
		log.Debugf("%v: %v", message, time.Since(start))
	}
}
