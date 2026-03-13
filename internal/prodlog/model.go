package prodlog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

var (
	logFolderPath atomic.Value
	disableStdout atomic.Bool
	logFile       *os.File
	mu            sync.Mutex
)

func DisableStdout() {
	disableStdout.Store(true)
}

func EnableLogFile(folderPath string) {
	logFolderPath.Store(folderPath)
}

func Info(v ...any) {
	output := fmt.Sprintln(v...)
	writeToFile(output)
	out(infLog, output)
}

func Infof(format string, v ...any) {
	output := fmt.Sprintf(format, v...)
	writeToFile(output)
	out(infLog, output)
}

func Warning(v ...any) {
	output := fmt.Sprintln(v...)
	writeToFile(output)
	out(warLog, output)
}

func Warningf(format string, v ...any) {
	output := fmt.Sprintf(format, v...)
	writeToFile(output)
	out(warLog, output)
}

func Error(v ...any) {
	output := fmt.Sprintln(v...)
	writeToFile(output)
	out(errLog, output)
}

func Errorf(format string, v ...any) {
	output := fmt.Sprintf(format, v...)
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

func writeToFile(content string) {
	if !hasLogFile() {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	filename := filepath.Join(logFolderPath.Load().(string), "log_"+time.Now().Format("060102"))
	if logFile == nil || logFile.Name() != filename {
		if logFile != nil {
			logFile.Close()
		}
		var err error
		logFile, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			errLog.Output(2, fmt.Sprintf("Failed to open new logfile: %v", err))
			return
		}
	}

	if _, err := logFile.WriteString(content); err != nil {
		errLog.Output(2, fmt.Sprintf("Failed to write to logfile: %v", err))
	}
}
