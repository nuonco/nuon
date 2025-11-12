package ui

import (
	"fmt"
	"os"
)

const logOutput string = "/tmp/.nuonctl.log"

func writeToLogFile(message string, args ...any) {
	if _, err := os.Stat(logOutput); os.IsNotExist(err) {
		file, err := os.Create(logOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to create file for log output")
			return
		}
		file.Close()
	}

	logFile, err := os.OpenFile(logOutput, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open file for log output")
		return
	}
	defer logFile.Close()

	fmt.Fprintf(logFile, message, args...)
}
