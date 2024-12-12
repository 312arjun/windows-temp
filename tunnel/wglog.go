package tunnel

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/windows/l18n"
)

func WGlog(format string, args ...interface{}) {

	// The main Eclipz log file on windows. There are two other vestigial log files
	// from WGW and eclipz for Linux.
	filename := "C:\\Program files\\Eclipz\\eclipz.log"

	pid := windows.GetCurrentProcessId()
	pidstr := strconv.Itoa(int(pid))

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		windows.MessageBox(0, windows.StringToUTF16Ptr("Must run Eclipz as Administrator"), windows.StringToUTF16Ptr(l18n.Sprintf("Error")), windows.MB_ICONERROR)
		os.Exit(1)
	}

	defer file.Close()

	now := time.Now()
	logtime := now.Format("2006-01-02 15:04:05.000")
	text := fmt.Sprintf(logtime+" WG ("+pidstr+") "+format+"\n", args...)
	if _, err = file.WriteString(text); err != nil {
		panic(err)
	}
}
