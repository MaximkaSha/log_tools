// Package utils provide common utils for agent and server.
package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"strconv"
)

// CheckIfStringIsNumber - return true if string is float64 or int64 value.
func CheckIfStringIsNumber(v string) bool {
	if _, err1 := strconv.Atoi(v); err1 == nil {
		return true
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return true
	}
	return false
}

// Float64ToByte - convert float64 value to bytes.Buffer.
func Float64ToByte(f float64) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, f)
	if err != nil {
		log.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

// CheckError - logs error.
func CheckError(err error) {
	if err != nil {
		log.Printf("error: %s", err)
	}
}

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
