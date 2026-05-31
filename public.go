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
	journal_ctl_fatal   = "<1>"
	journal_ctl_error   = "<3>"
	journal_ctl_warning = "<4>"
	journal_ctl_info    = "<6>"
	outLogPreStr        = journal_ctl_info + "INFO: "
	warLogPreStr        = journal_ctl_warning + "WARNING: "
	errLogPreStr        = journal_ctl_error + "ERROR: "
	fatalLogPreStr      = journal_ctl_fatal + "FATAL ERROR: "
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
	output := fmt.Sprintln(v...)
	writeToFile(outLogPreStr, output)
	out(infLog, 0, output)
}

// InfoDepth behaves like Info but adds skip extra stack frames when reporting
// the source file and line, so a logging wrapper can have its own caller
// reported instead of the wrapper itself. Direct callers should use Info.
func InfoDepth(skip int, v ...any) {
	output := fmt.Sprintln(v...)
	writeToFile(outLogPreStr, output)
	out(infLog, skip, output)
}

func Infof(format string, v ...any) {
	output := fmt.Sprintf(format, v...)
	writeToFile(outLogPreStr, output)
	out(infLog, 0, output)
}

// InfofDepth behaves like Infof but adds skip extra stack frames when reporting
// the source file and line.
func InfofDepth(skip int, format string, v ...any) {
	output := fmt.Sprintf(format, v...)
	writeToFile(outLogPreStr, output)
	out(infLog, skip, output)
}

func Warning(v ...any) {
	output := fmt.Sprintln(v...)
	writeToFile(warLogPreStr, output)
	out(warLog, 0, output)
}

// WarningDepth behaves like Warning but adds skip extra stack frames when
// reporting the source file and line.
func WarningDepth(skip int, v ...any) {
	output := fmt.Sprintln(v...)
	writeToFile(warLogPreStr, output)
	out(warLog, skip, output)
}

func Warningf(format string, v ...any) {
	output := fmt.Sprintf(format, v...)
	writeToFile(warLogPreStr, output)
	out(warLog, 0, output)
}

// WarningfDepth behaves like Warningf but adds skip extra stack frames when
// reporting the source file and line.
func WarningfDepth(skip int, format string, v ...any) {
	output := fmt.Sprintf(format, v...)
	writeToFile(warLogPreStr, output)
	out(warLog, skip, output)
}

func Error(v ...any) {
	output := fmt.Sprintln(v...)
	writeToFile(errLogPreStr, output)
	out(errLog, 0, output)
}

// ErrorDepth behaves like Error but adds skip extra stack frames when reporting
// the source file and line.
func ErrorDepth(skip int, v ...any) {
	output := fmt.Sprintln(v...)
	writeToFile(errLogPreStr, output)
	out(errLog, skip, output)
}

func Errorf(format string, v ...any) {
	output := fmt.Sprintf(format, v...)
	writeToFile(errLogPreStr, output)
	out(errLog, 0, output)
}

// ErrorfDepth behaves like Errorf but adds skip extra stack frames when
// reporting the source file and line.
func ErrorfDepth(skip int, format string, v ...any) {
	output := fmt.Sprintf(format, v...)
	writeToFile(errLogPreStr, output)
	out(errLog, skip, output)
}

func Fatal(v ...any) {
	fatal(0, fmt.Sprintln(v...))
}

// FatalDepth behaves like Fatal but adds skip extra stack frames when reporting
// the source file and line.
func FatalDepth(skip int, v ...any) {
	fatal(skip, fmt.Sprintln(v...))
}

func Fatalf(format string, v ...any) {
	fatal(0, fmt.Sprintf(format, v...))
}

// FatalfDepth behaves like Fatalf but adds skip extra stack frames when
// reporting the source file and line.
func FatalfDepth(skip int, format string, v ...any) {
	fatal(skip, fmt.Sprintf(format, v...))
}

func fatal(skip int, output string) {
	writeToFile(fatalLogPreStr, output)
	errLog.Output(3+skip, output)
	os.Exit(1)
}

func hasLogFile() bool {
	path, ok := logFolderPath.Load().(string)
	return ok && len(path) > 0
}

func out(logger *log.Logger, skip int, content string) {
	if disableStdout.Load() {
		return
	}
	logger.Output(3+skip, content)
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

func writeToFile(prefix, content string) {
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
	if _, err := fmt.Fprintf(logFile, "%s: %s %s %s", time.Now().Format("15:04:05"), prefix, content, newline); err != nil {
		errLog.Output(2, fmt.Sprintf("Failed to write to logfile: %v", err))
	}
}
