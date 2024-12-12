package eclipz

import (
	"fmt"
	"log"
	"os"
)

func OpenLogFile() {
	filename := Config.Client.LogFile
	if filename == "" {
		filename = getLogFileName()
	}

	// open log file
	logFile, err := os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Unable to open %s for logging\n", filename)
		return
	}

	// Set log out put and enjoy :)
	log.SetOutput(logFile)

	// optional: log date-time, filename, and line number
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}
