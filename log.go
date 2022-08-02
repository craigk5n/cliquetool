package main

import "fmt"

var loggingLevel int = 0

func logInit(path string, verboseLevel int) {
	loggingLevel = verboseLevel
}

func logMessage(message string) {
	// TODO: Write to file
	if loggingLevel > 0 {
		fmt.Println(message)
	}
}
