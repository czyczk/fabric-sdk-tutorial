package timingutils

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// GetDeferrableTimingLogger creates a logger function that starts a timer when called and ends the timer when the calling function ends and logs (at debug level) the time diff.
func GetDeferrableTimingLogger(message string) func() {
	if !global.ShowTimingLogs {
		return func() {}
	}

	start := time.Now()
	return func() {
		log.Debugf("%v: %v", message, time.Since(start))
	}
}

// getFileDescriptorAppendMode gets a file descriptor for the specified file path as append mode. Useful for appending logs to the file.
func getFileDescriptorAppendMode(filename string) (f *os.File, err error) {
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

func writeLineToFile(str string, f *os.File) error {
	if f == nil {
		return fmt.Errorf("文件描述符为 nil")
	}

	if _, err := f.WriteString(str + "\n"); err != nil {
		return errors.Wrapf(err, "无法往文件 %v 中添加内容", f.Name())
	}

	return nil
}

// This is a file logger that contains functions to append timestamps to the specified log files. It can append the start time of an action to the start log file, or the complete / error time of the action to the end log file. The timestamp logs are formatted as lines of "${InstanceID}~${timestamp}~${isSuccess}".
type StartEndFileLogger struct {
	InstanceID   string
	StartLogFile *os.File
	EndLogFile   *os.File
}

// NewStartEndFileLogger creates a StartEndFileLogger instance providing with the file paths to the log files.
func NewStartEndFileLogger(instanceID string, startLogFilename, endLogFilename string) (*StartEndFileLogger, error) {
	startLogFd, err := getFileDescriptorAppendMode(startLogFilename)
	if err != nil {
		return nil, err
	}

	endLogFd, err := getFileDescriptorAppendMode(endLogFilename)
	if err != nil {
		return nil, err
	}

	return NewStartEndFileLoggerWithFileDescriptors(instanceID, startLogFd, endLogFd), nil
}

// NewStartEndFileLoggerWithFileDescriptors creates a StartEndFileLogger instance providing with appropriate opened file descriptors.
func NewStartEndFileLoggerWithFileDescriptors(instanceID string, startLogFileDescriptor, endLogFileDescriptor *os.File) *StartEndFileLogger {
	return &StartEndFileLogger{
		InstanceID:   instanceID,
		StartLogFile: startLogFileDescriptor,
		EndLogFile:   endLogFileDescriptor,
	}
}

func (l *StartEndFileLogger) getFormattedLine(timestamp time.Time, isSuccess bool) (string, error) {
	isSuccessStr := "T"
	if !isSuccess {
		isSuccessStr = "F"
	}

	timestampStr, err := SerializeTimestamp(timestamp)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v~%v~%v", l.InstanceID, timestampStr, isSuccessStr), nil
}

func (l *StartEndFileLogger) logTimestampToFile(timestamp time.Time, isSuccess bool, targetFile *os.File) error {
	line, err := l.getFormattedLine(timestamp, isSuccess)
	if err != nil {
		return err
	}

	err = writeLineToFile(line, targetFile)
	if err != nil {
		return err
	}

	return nil
}

// LogStart logs a formatted timestamp of now to the strat log file. It should be called right before an action is started.
func (l *StartEndFileLogger) LogStart() error {
	return l.LogStartWithTimestamp(time.Now())
}

// LogStartWithTimestamp logs a formatted line of the specified timestamp to the start log file. The start timestamp should be collected right before the action is started and this function can be called anytime.
func (l *StartEndFileLogger) LogStartWithTimestamp(timestamp time.Time) error {
	err := l.logTimestampToFile(timestamp, true, l.StartLogFile)
	if err != nil {
		return err
	}

	return nil
}

// LogSuccess logs a formatted timestamp of now with an action success status of `true` to the end log file. It should be called right after the action is successful.
func (l *StartEndFileLogger) LogSuccess() error {
	return l.LogSuccessWithTimestamp(time.Now())
}

// LogSuccessWithTimestamp logs a formatted line of the specified timestamp with an action success status of `true` to the end log file. The end timestamp should be collected right after the action is successful and this function can be called anytime.
func (l *StartEndFileLogger) LogSuccessWithTimestamp(timestamp time.Time) error {
	err := l.logTimestampToFile(timestamp, true, l.EndLogFile)
	if err != nil {
		return err
	}

	return nil
}

// LogFailure logs a formatted timestamp of now with an action success status of `false` to the end log file. It should be called right after the action fails.
func (l *StartEndFileLogger) LogFailure() error {
	return l.LogFailureWithTimestamp(time.Now())
}

// LogFailurewithTimestamp logs a formatted line of the specified timestamp with an action success status of `false` to the end log file. The end timestamp shoukld be collected right after the action fails and this functino can be called anytime.
func (l *StartEndFileLogger) LogFailureWithTimestamp(timestamp time.Time) error {
	err := l.logTimestampToFile(timestamp, false, l.EndLogFile)
	if err != nil {
		return err
	}

	return nil
}

func (l *StartEndFileLogger) Close() (errs []error) {
	mu := sync.RWMutex{}

	defer func() {
		err := l.StartLogFile.Close()
		if err != nil {
			mu.Lock()
			defer mu.Unlock()
			errs = append(errs, err)
		}
	}()

	defer func() {
		err := l.EndLogFile.Close()
		if err != nil {
			mu.Lock()
			defer mu.Unlock()
			errs = append(errs, err)
		}
	}()

	return
}
