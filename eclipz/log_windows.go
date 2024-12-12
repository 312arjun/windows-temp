package eclipz

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/sys/windows"
)

/* Logging utility to add output into the Eclipz log we keep in WGW.
 * The log file & path should be the same as the one used by Eclipz's WireGuard-Windows.
 */

func EClog(format string, args ...interface{}) {
	// Config.Client.LogFile is not set. In fact the object Config.Client seems
	// to be exactly whats in the config.json file.
	//filename := Config.Client.LogFile
	filename := "C:\\Program files\\Eclipz\\eclipz.log"

	pid := windows.GetCurrentProcessId()
	pidstr := strconv.Itoa(int(pid))

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	now := time.Now()
	logtime := now.Format("2006-01-02 15:04:05.000")
	text := fmt.Sprintf(logtime+" EC ("+pidstr+") "+format+"\n", args...)
	if _, err = file.WriteString(text); err != nil {
		panic(err)
	}
}
