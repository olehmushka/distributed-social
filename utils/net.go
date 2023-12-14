package utils

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/olehmushka/distributed-social/utils/stringutils"
)

func ExtractIPAddress(r *http.Request) (string, error) {
	// Get the IP address from the "X-Forwarded-For" header.
	ipAddress := r.Header.Get("X-Forwarded-For")

	if ipAddress == "" {
		// If the "X-Forwarded-For" header is not set, try to get the IP address from the remote address of the connection.
		ipAddress, _, err := SplitHostPort(r.RemoteAddr)
		if err != nil {
			return "", fmt.Errorf("failed extract ip addr from remote addr (addr=%v)", r.RemoteAddr)
		}

		return ipAddress, nil
	}

	// If the "X-Forwarded-For" header is set, split it and return the first IP address in the list.
	ipAddresses := strings.Split(ipAddress, ", ")
	if len(ipAddresses) < 1 {
		return "", fmt.Errorf("failed extract ip addr from forwarded (addr=%v)", ipAddress)
	}

	return ipAddresses[0], nil
}

func SplitHostPort(hostport string) (host string, port int, err error) {
	var portStr string
	host, portStr, err = net.SplitHostPort(hostport)
	if err != nil {
		return "", 0, err
	}

	var portUint uint64
	portUint, err = stringutils.StringToUInt64(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("failed parse port (port=%v)", portStr)
	}

	return host, int(portUint), nil
}
