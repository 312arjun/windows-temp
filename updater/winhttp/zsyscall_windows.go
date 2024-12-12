// Code generated by 'go generate'; DO NOT EDIT.

package winhttp

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer

// Do the interface allocations only once for common
// Errno values.
const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
	errERROR_EINVAL     error = syscall.EINVAL
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return errERROR_EINVAL
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

var (
	modwinhttp = windows.NewLazySystemDLL("winhttp.dll")

	procWinHttpCloseHandle        = modwinhttp.NewProc("WinHttpCloseHandle")
	procWinHttpConnect            = modwinhttp.NewProc("WinHttpConnect")
	procWinHttpCrackUrl           = modwinhttp.NewProc("WinHttpCrackUrl")
	procWinHttpOpen               = modwinhttp.NewProc("WinHttpOpen")
	procWinHttpOpenRequest        = modwinhttp.NewProc("WinHttpOpenRequest")
	procWinHttpQueryDataAvailable = modwinhttp.NewProc("WinHttpQueryDataAvailable")
	procWinHttpQueryHeaders       = modwinhttp.NewProc("WinHttpQueryHeaders")
	procWinHttpReadData           = modwinhttp.NewProc("WinHttpReadData")
	procWinHttpReceiveResponse    = modwinhttp.NewProc("WinHttpReceiveResponse")
	procWinHttpSendRequest        = modwinhttp.NewProc("WinHttpSendRequest")
	procWinHttpSetOption          = modwinhttp.NewProc("WinHttpSetOption")
	procWinHttpSetStatusCallback  = modwinhttp.NewProc("WinHttpSetStatusCallback")
)

func winHttpCloseHandle(handle _HINTERNET) (err error) {
	r1, _, e1 := syscall.Syscall(procWinHttpCloseHandle.Addr(), 1, uintptr(handle), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func winHttpConnect(sessionHandle _HINTERNET, serverName *uint16, serverPort uint16, reserved uint32) (handle _HINTERNET, err error) {
	r0, _, e1 := syscall.Syscall6(procWinHttpConnect.Addr(), 4, uintptr(sessionHandle), uintptr(unsafe.Pointer(serverName)), uintptr(serverPort), uintptr(reserved), 0, 0)
	handle = _HINTERNET(r0)
	if handle == 0 {
		err = errnoErr(e1)
	}
	return
}

func winHttpCrackUrl(url *uint16, urlSize uint32, flags uint32, components *_URL_COMPONENTS) (err error) {
	r1, _, e1 := syscall.Syscall6(procWinHttpCrackUrl.Addr(), 4, uintptr(unsafe.Pointer(url)), uintptr(urlSize), uintptr(flags), uintptr(unsafe.Pointer(components)), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func winHttpOpen(userAgent *uint16, accessType uint32, proxy *uint16, proxyBypass *uint16, flags uint32) (sessionHandle _HINTERNET, err error) {
	r0, _, e1 := syscall.Syscall6(procWinHttpOpen.Addr(), 5, uintptr(unsafe.Pointer(userAgent)), uintptr(accessType), uintptr(unsafe.Pointer(proxy)), uintptr(unsafe.Pointer(proxyBypass)), uintptr(flags), 0)
	sessionHandle = _HINTERNET(r0)
	if sessionHandle == 0 {
		err = errnoErr(e1)
	}
	return
}

func winHttpOpenRequest(connectHandle _HINTERNET, verb *uint16, objectName *uint16, version *uint16, referrer *uint16, acceptTypes **uint16, flags uint32) (requestHandle _HINTERNET, err error) {
	r0, _, e1 := syscall.Syscall9(procWinHttpOpenRequest.Addr(), 7, uintptr(connectHandle), uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(objectName)), uintptr(unsafe.Pointer(version)), uintptr(unsafe.Pointer(referrer)), uintptr(unsafe.Pointer(acceptTypes)), uintptr(flags), 0, 0)
	requestHandle = _HINTERNET(r0)
	if requestHandle == 0 {
		err = errnoErr(e1)
	}
	return
}

func winHttpQueryDataAvailable(requestHandle _HINTERNET, bytesAvailable *uint32) (err error) {
	r1, _, e1 := syscall.Syscall(procWinHttpQueryDataAvailable.Addr(), 2, uintptr(requestHandle), uintptr(unsafe.Pointer(bytesAvailable)), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func winHttpQueryHeaders(requestHandle _HINTERNET, infoLevel uint32, name *uint16, buffer unsafe.Pointer, bufferLen *uint32, index *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procWinHttpQueryHeaders.Addr(), 6, uintptr(requestHandle), uintptr(infoLevel), uintptr(unsafe.Pointer(name)), uintptr(buffer), uintptr(unsafe.Pointer(bufferLen)), uintptr(unsafe.Pointer(index)))
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func winHttpReadData(requestHandle _HINTERNET, buffer *byte, bufferSize uint32, bytesRead *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procWinHttpReadData.Addr(), 4, uintptr(requestHandle), uintptr(unsafe.Pointer(buffer)), uintptr(bufferSize), uintptr(unsafe.Pointer(bytesRead)), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func winHttpReceiveResponse(requestHandle _HINTERNET, reserved uintptr) (err error) {
	r1, _, e1 := syscall.Syscall(procWinHttpReceiveResponse.Addr(), 2, uintptr(requestHandle), uintptr(reserved), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func winHttpSendRequest(requestHandle _HINTERNET, headers *uint16, headersLength uint32, optional *byte, optionalLength uint32, totalLength uint32, context uintptr) (err error) {
	r1, _, e1 := syscall.Syscall9(procWinHttpSendRequest.Addr(), 7, uintptr(requestHandle), uintptr(unsafe.Pointer(headers)), uintptr(headersLength), uintptr(unsafe.Pointer(optional)), uintptr(optionalLength), uintptr(totalLength), uintptr(context), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func winHttpSetOption(sessionOrRequestHandle _HINTERNET, option uint32, buffer unsafe.Pointer, bufferLen uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procWinHttpSetOption.Addr(), 4, uintptr(sessionOrRequestHandle), uintptr(option), uintptr(buffer), uintptr(bufferLen), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func winHttpSetStatusCallback(handle _HINTERNET, callback uintptr, notificationFlags uint32, reserved uintptr) (previousCallback uintptr, err error) {
	r0, _, e1 := syscall.Syscall6(procWinHttpSetStatusCallback.Addr(), 4, uintptr(handle), uintptr(callback), uintptr(notificationFlags), uintptr(reserved), 0, 0)
	previousCallback = uintptr(r0)
	if previousCallback == _WINHTTP_INVALID_STATUS_CALLBACK {
		err = errnoErr(e1)
	}
	return
}
