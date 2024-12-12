package eclipz

import (
	"net"
	"fmt"
	"strings"
	"strconv"
	"encoding/hex"
	"encoding/base64"
)

func Get16(data []byte) uint16 {
	var n uint16
	n = uint16(data[0]) << 8 | uint16(data[1])
	return n
}

func Put16(data []byte, n uint16) {
	data[0] = byte((n>>8) & 0xff)
	data[1] = byte(n & 0xff)
}

func Get32(data []byte) uint32 {
	var n uint32
	n = uint32(data[0]) << 24 | uint32(data[1]) << 16 | uint32(data[2]) << 8 | uint32(data[3])
	return n
}

func Get16i(data []byte) int {
	return int(Get16(data))
}

func Get32i(data []byte) int {
	return int(Get32(data))
}

func Put32(data []byte, n uint32) {
	data[0] = byte((n>>24) & 0xff)
	data[1] = byte((n>>16) & 0xff)
	data[2] = byte((n>>8) & 0xff)
	data[3] = byte(n & 0xff)
}

func Get64(data []byte) uint64 {
	var n uint64
	n = uint64(data[0]) << 56 | uint64(data[1]) << 48 | uint64(data[2]) << 40 | uint64(data[3]) << 32 |
	    uint64(data[0]) << 24 | uint64(data[1]) << 16 | uint64(data[2]) << 8 | uint64(data[3])
	return n
}

func Put64(data []byte, n uint64) {
	data[0] = byte((n>>56) & 0xff)
	data[1] = byte((n>>48) & 0xff)
	data[2] = byte((n>>40) & 0xff)
	data[3] = byte((n>>32) & 0xff)

	data[4] = byte((n>>24) & 0xff)
	data[5] = byte((n>>16) & 0xff)
	data[6] = byte((n>>8) & 0xff)
	data[7] = byte(n & 0xff)
}

func Str2UDPAddr(addrStr string) (*net.UDPAddr, bool) {
	udpAddr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return nil, false
	}
	return udpAddr, true
}

func HEX(data []byte) string {
	return hex.EncodeToString(data)
}

// Truncated String
func Truncate(s string) string {
	n := len(s)
	if n > 12 {
		return s[0:4] + "..." + s[n-4:n]
	}
	return s
}

// Truncated Hex
func HEXT(data []byte) string {
	n := len(data)
	if n > 16 {
		return fmt.Sprintf("%02x%02x%02x...%02x%02x%02x (%d bytes)",
			data[0], data[1], data[2],
			data[n-3], data[n-2], data[n-1], n)
	}
	return hex.EncodeToString(data)
}

func Base64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func Base64decode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

func AddressWithPort(addrStr string, defPort int) (string, int) {
	s := strings.Split(addrStr, ":")
	switch len(s) {
	case 0:
		// empty string
	case 1:
		// port not specified
	case 2:
		// IPv4 name or addresss and port specified
		p, _ := strconv.Atoi(s[1])
		if p == 0 {
			p = defPort
		}
		return s[0], p
	default:
		// possibly an IPv6 address
	}
	return addrStr, defPort
}
