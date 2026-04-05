package prodlog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	journal_ctl_error   = "<3>"
	journal_ctl_warning = "<4>"
	journal_ctl_info    = "<6>"
	outLogPreStr        = journal_ctl_info + "INFO: "
	warLogPreStr        = journal_ctl_warning + "WARNING: "
	errLogPreStr        = journal_ctl_error + "ERROR: "
)

var (
	infLog = log.New(os.Stdout, outLogPreStr, log.LstdFlags|log.Lshortfile)
	warLog = log.New(os.Stdout, warLogPreStr, log.LstdFlags|log.Lshortfile)
	errLog = log.New(os.Stderr, errLogPreStr, log.LstdFlags|log.Llongfile)
)

const log_file_prefix = "log"

var (
	customLogFilePrefix atomic.Value
	logFolderPath       atomic.Value
	disableStdout       atomic.Bool
	logFile             *os.File
	mu                  sync.Mutex
	newline             string = "\n"
)

func DisableStdout() {
	disableStdout.Store(true)
}

func EnableLogFile(folderPath string) {
	if runtime.GOOS == "windows" {
		newline = "\r\n"
	}
	logFolderPath.Store(folderPath)
}

func SetLogFilePrefix(prefix string) {
	customLogFilePrefix.Store(prefix)
}

func Info(v ...any) {
	output := outLogPreStr + fmt.Sprintln(v...)
	writeToFile(output)
	out(infLog, output)
}

func Infof(format string, v ...any) {
	output := outLogPreStr + fmt.Sprintf(format, v...)
	writeToFile(output)
	out(infLog, output)
}

func Warning(v ...any) {
	output := warLogPreStr + fmt.Sprintln(v...)
	writeToFile(output)
	out(warLog, output)
}

func Warningf(format string, v ...any) {
	output := warLogPreStr + fmt.Sprintf(format, v...)
	writeToFile(output)
	out(warLog, output)
}

func Error(v ...any) {
	output := errLogPreStr + fmt.Sprintln(v...)
	writeToFile(output)
	out(errLog, output)
}

func Errorf(format string, v ...any) {
	output := errLogPreStr + fmt.Sprintf(format, v...)
	writeToFile(output)
	out(errLog, output)
}

func Fatal(v ...any) {
	output := fmt.Sprintln(v...)
	writeToFile(output)
	errLog.Output(2, output)
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	output := fmt.Sprintf(format, v...)
	writeToFile(output)
	errLog.Output(2, output)

	os.Exit(1)
}

func hasLogFile() bool {
	path, ok := logFolderPath.Load().(string)
	return ok && len(path) > 0
}

func out(logger *log.Logger, content string) {
	if disableStdout.Load() {
		return
	}
	logger.Output(3, content)
}

func getLogFilePrefix() string {
	prefix := log_file_prefix
	if customPrefix, ok := customLogFilePrefix.Load().(string); ok && len(customPrefix) > 0 {
		prefix = customPrefix
	}
	return prefix
}

func getFullLogFilePath() string {
	filename := getLogFilePrefix() + "_" + time.Now().Format("060102") + ".log"
	path := filepath.Join(logFolderPath.Load().(string), filename)
	return path
}

func writeToFile(content string) {
	if !hasLogFile() {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	path := getFullLogFilePath()
	if logFile == nil || logFile.Name() != path {
		if logFile != nil {
			logFile.Close()
		}
		var err error
		logFile, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			errLog.Output(2, fmt.Sprintf("Failed to open new logfile: %v", err))
			return
		}
	}
	if _, err := logFile.WriteString(content + newline); err != nil {
		errLog.Output(2, fmt.Sprintf("Failed to write to logfile: %v", err))
	}
}
