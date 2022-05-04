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
	wg           *sync.WaitGroup
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

	wg := sync.WaitGroup{}

	return NewStartEndFileLoggerWithFileDescriptors(instanceID, startLogFd, endLogFd, &wg), nil
}

// NewStartEndFileLoggerWithFileDescriptors creates a StartEndFileLogger instance providing with appropriate opened file descriptors.
func NewStartEndFileLoggerWithFileDescriptors(instanceID string, startLogFileDescriptor, endLogFileDescriptor *os.File, wg *sync.WaitGroup) *StartEndFileLogger {
	return &StartEndFileLogger{
		InstanceID:   instanceID,
		StartLogFile: startLogFileDescriptor,
		EndLogFile:   endLogFileDescriptor,
		wg:           wg,
	}
}

func (l *StartEndFileLogger) getFormattedLine(taskID string, timestamp time.Time, isSuccess bool) (string, error) {
	isSuccessStr := "T"
	if !isSuccess {
		isSuccessStr = "F"
	}

	timestampStr, err := SerializeTimestamp(timestamp)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v~%v~%v~%v", l.InstanceID, taskID, timestampStr, isSuccessStr), nil
}

func (l *StartEndFileLogger) logTimestampToFile(taskID string, timestamp time.Time, isSuccess bool, targetFile *os.File) error {
	line, err := l.getFormattedLine(taskID, timestamp, isSuccess)
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
func (l *StartEndFileLogger) LogStart(taskID string) error {
	return l.LogStartWithTimestamp(taskID, time.Now())
}

// LogStartAsync is the async version of LogStart.
func (l *StartEndFileLogger) LogStartAsync(taskID string, chanErr chan<- error) {
	l.wg.Add(1)
	go func() {
		err := l.LogStart(taskID)
		chanErr <- err
		l.wg.Done()
	}()
}

// LogStartWithTimestamp logs a formatted line of the specified timestamp to the start log file. The start timestamp should be collected right before the action is started and this function can be called anytime.
func (l *StartEndFileLogger) LogStartWithTimestamp(taskID string, timestamp time.Time) error {
	err := l.logTimestampToFile(taskID, timestamp, true, l.StartLogFile)
	if err != nil {
		return err
	}

	return nil
}

// LogStartWithTimestampAsync is the async version of LogStartWithTimestamp.
func (l *StartEndFileLogger) LogStartWithTimestampAsync(taskID string, timestamp time.Time, chanErr chan<- error) {
	l.wg.Add(1)
	go func() {
		err := l.LogStartWithTimestamp(taskID, timestamp)
		chanErr <- err
		l.wg.Done()
	}()
}

// LogSuccess logs a formatted timestamp of now with an action success status of `true` to the end log file. It should be called right after the action is successful.
func (l *StartEndFileLogger) LogSuccess(taskID string) error {
	return l.LogSuccessWithTimestamp(taskID, time.Now())
}

// LogSuccessAsync is the async version of LogSuccess.
func (l *StartEndFileLogger) LogSuccessAsync(taskID string, chanErr chan<- error) {
	l.wg.Add(1)
	go func() {
		err := l.LogSuccess(taskID)
		chanErr <- err
		l.wg.Done()
	}()
}

// LogSuccessWithTimestamp logs a formatted line of the specified timestamp with an action success status of `true` to the end log file. The end timestamp should be collected right after the action is successful and this function can be called anytime.
func (l *StartEndFileLogger) LogSuccessWithTimestamp(taskID string, timestamp time.Time) error {
	err := l.logTimestampToFile(taskID, timestamp, true, l.EndLogFile)
	if err != nil {
		return err
	}

	return nil
}

// LogSuccessWithTimestampAsync is the async version of LogSuccessWithTimestamp
func (l *StartEndFileLogger) LogSuccessWithTimestampAsync(taskID string, timestamp time.Time, chanErr chan<- error) {
	l.wg.Add(1)
	go func() {
		err := l.LogSuccessWithTimestamp(taskID, timestamp)
		chanErr <- err
		l.wg.Done()
	}()
}

// LogFailure logs a formatted timestamp of now with an action success status of `false` to the end log file. It should be called right after the action fails.
func (l *StartEndFileLogger) LogFailure(taskID string) error {
	return l.LogFailureWithTimestamp(taskID, time.Now())
}

// LogFailureAsync is the async version of LogFailure
func (l *StartEndFileLogger) LogFailureAsync(taskID string, chanErr chan<- error) {
	l.wg.Add(1)
	go func() {
		err := l.LogFailure(taskID)
		chanErr <- err
		l.wg.Done()
	}()
}

// LogFailurewithTimestamp logs a formatted line of the specified timestamp with an action success status of `false` to the end log file. The end timestamp shoukld be collected right after the action fails and this functino can be called anytime.
func (l *StartEndFileLogger) LogFailureWithTimestamp(taskID string, timestamp time.Time) error {
	err := l.logTimestampToFile(taskID, timestamp, false, l.EndLogFile)
	if err != nil {
		return err
	}

	return nil
}

// LogFailureWithTimestampAsync is the async version of LogFailureWithTimestamp
func (l *StartEndFileLogger) LogFailureWithTimestampAsync(taskID string, timestamp time.Time, chanErr chan<- error) {
	l.wg.Add(1)
	go func() {
		err := l.LogFailureWithTimestamp(taskID, timestamp)
		chanErr <- err
		l.wg.Done()
	}()
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

	l.wg.Wait()

	return
}
