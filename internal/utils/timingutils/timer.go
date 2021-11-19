package timingutils

import (
	"fmt"
	"os"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"github.com/pkg/errors"
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

func GetFileDescriptorAppendMode(filename string) (f *os.File, err error) {
	f, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		err = errors.Wrapf(err, "无法以附加模式打开文件 %v")
		return
	}

	return
}

func SerializeTimestamp(timestamp time.Time) (timestampStr string, err error) {
	timestampBytes, err := timestamp.MarshalText()
	if err != nil {
		err = errors.Wrap(err, "无法序列化时间戳")
		return
	}

	timestampStr = string(timestampBytes)
	return
}

func SerializeDuration(duration time.Duration) string {
	return duration.String()
}

func WriteStringToFile(str string, f *os.File) error {
	if f == nil {
		return fmt.Errorf("文件描述符为 nil")
	}

	if _, err := f.WriteString(str + "\n"); err != nil {
		return errors.Wrapf(err, "无法往文件 %v 中添加内容", f.Name())
	}

	return nil
}
