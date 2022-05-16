package helper

import (
	"fmt"
	"os"
	"testing"
)

func TestCreateLogFile(t *testing.T) {
	logFile := fmt.Sprintf("%s.log", os.Args[0])
	LogInit(logFile, 0)

	defer LogRelease()
	defer os.Remove(logFile)

	if _, err := os.Stat(logFile); err != nil {
		t.Error("log file not exists")
	}

}
