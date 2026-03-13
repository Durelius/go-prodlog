package prodlog_test

import (
	"github.com/durelius/go-prodlog/internal/prodlog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func todayString() string {
	return time.Now().Format("060102")
}

func logPath(dir string) string {
	return filepath.Join(dir, "log_"+todayString())
}

func readLog(t *testing.T, dir string) string {
	t.Helper()
	content, err := os.ReadFile(logPath(dir))
	if err != nil {
		t.Fatalf("could not read log file: %v", err)
	}
	return string(content)
}

func TestInfo_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	prodlog.EnableLogFile(dir)
	prodlog.Info("info message")
	if !strings.Contains(readLog(t, dir), "info message") {
		t.Error("expected info message in log file")
	}
}

func TestInfof_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	prodlog.EnableLogFile(dir)
	prodlog.Infof("hello %s %d", "world", 42)
	if !strings.Contains(readLog(t, dir), "hello world 42") {
		t.Error("expected formatted info message in log file")
	}
}

func TestWarning_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	prodlog.EnableLogFile(dir)
	prodlog.Warning("warn message")
	if !strings.Contains(readLog(t, dir), "warn message") {
		t.Error("expected warning in log file")
	}
}

func TestError_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	prodlog.EnableLogFile(dir)
	prodlog.Error("error message")
	if !strings.Contains(readLog(t, dir), "error message") {
		t.Error("expected error in log file")
	}
}

func TestAllLevels_WriteToFile(t *testing.T) {
	dir := t.TempDir()
	prodlog.EnableLogFile(dir)
	prodlog.Info("info line")
	prodlog.Warning("warning line")
	prodlog.Error("error line")

	content := readLog(t, dir)
	for _, expected := range []string{"info line", "warning line", "error line"} {
		if !strings.Contains(content, expected) {
			t.Errorf("expected %q in log file", expected)
		}
	}
}

func TestNoFile_WithoutEnableLogFile(t *testing.T) {
	dir := t.TempDir()
	prodlog.Info("no file")
	if _, err := os.Stat(logPath(dir)); !os.IsNotExist(err) {
		t.Error("expected no log file when EnableLogFile not called")
	}
}

func TestAppend_ToExistingFile(t *testing.T) {
	dir := t.TempDir()
	prodlog.EnableLogFile(dir)
	prodlog.Info("first")
	prodlog.EnableLogFile(dir)
	prodlog.Info("second")

	content := readLog(t, dir)
	if !strings.Contains(content, "first") || !strings.Contains(content, "second") {
		t.Error("expected both messages in file after re-enable")
	}
}

func TestConcurrent_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	prodlog.EnableLogFile(dir)

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			prodlog.Infof("goroutine %d", n)
		}(i)
	}
	wg.Wait()

	content := readLog(t, dir)
	if len(content) == 0 {
		t.Error("expected log file to have content after concurrent writes")
	}
}
