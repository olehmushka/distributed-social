package netutils

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/olehmushka/distributed-social/utils/stringutils"
)

// CloneURL returns a deep clone of u.  The User pointer of clone is the same,
// since a *url.Userinfo is effectively an immutable value.
func CloneURL(u *url.URL) (clone *url.URL) {
	if u == nil {
		return nil
	}

	cloneVal := *u

	return &cloneVal
}

// IsValidHostInnerRune returns true if r is a valid inner—that is, neither
// initial nor final—rune for a hostname label.
func IsValidHostInnerRune(r rune) (ok bool) {
	return r == '-' || IsValidHostOuterRune(r)
}

// IsValidHostOuterRune returns true if r is a valid initial or final rune for
// a hostname label.
func IsValidHostOuterRune(r rune) (ok bool) {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

// JoinHostPort is a convenient wrapper for net.JoinHostPort with port of type
// int.  As opposed to net.JoinHostPort it also trims the host from square
// brackets if any.  This may be useful when passing url.URL.Host field
// containing an IPv6 address.
func JoinHostPort(host string, port int) (hostport string) {
	return net.JoinHostPort(strings.Trim(host, "[]"), strconv.Itoa(port))
}

// SplitHostPort is a convenient wrapper for [net.SplitHostPort] with port of
// type int.
func SplitHostPort(hostport string) (host string, port int, err error) {
	var portStr string
	host, portStr, err = net.SplitHostPort(hostport)
	if err != nil {
		return "", 0, err
	}

	var portUint uint64
	portUint, err = stringutils.StringToUInt64(portStr)
	if err != nil {
		return "", 0, err
	}

	return host, int(portUint), nil
}

// SplitHost is a wrapper for [net.SplitHostPort] for cases when the hostport
// may or may not contain a port.
func SplitHost(hostport string) (host string, err error) {
	host, _, err = net.SplitHostPort(hostport)
	if err != nil {
		// Check for the missing port error.  If it is that error, just
		// use the host as is.
		//
		// See the source code for net.SplitHostPort.
		const missingPort = "missing port in address"

		addrErr := &net.AddrError{}
		if !errors.As(err, &addrErr) || addrErr.Err != missingPort {
			return "", err
		}

		host = hostport
	}

	return host, nil
}
