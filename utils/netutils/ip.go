package netutils

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/olehmushka/distributed-social/utils/sliceutils"
)

func ExtractIPAddress(r *http.Request) (string, error) {
	// Get the IP address from the "X-Forwarded-For" header.
	ipAddress := r.Header.Get("X-Forwarded-For")

	if ipAddress == "" {
		// If the "X-Forwarded-For" header is not set, try to get the IP address from the remote address of the connection.
		ipAddress, _, err := SplitHostPort(r.RemoteAddr)
		if err != nil {
			return "", err
		}

		return ipAddress, nil
	}

	// If the "X-Forwarded-For" header is set, split it and return the first IP address in the list.
	ipAddresses := strings.Split(ipAddress, ", ")
	if len(ipAddresses) < 1 {
		return "", fmt.Errorf("failed extract IP address from forwarded (IP address=%v)", ipAddress)
	}

	return ipAddresses[0], nil
}

// Bit lengths of IP addresses.
const (
	IPv4BitLen = net.IPv4len * 8
	IPv6BitLen = net.IPv6len * 8
)

func CloneIPs(ips []net.IP) (clone []net.IP) {
	if ips == nil {
		return nil
	}

	clone = make([]net.IP, len(ips))
	for i, ip := range ips {
		clone[i] = sliceutils.Clone(ip)
	}

	return clone
}

func IPAndPortFromAddr(addr net.Addr) (ip net.IP, port int) {
	switch addr := addr.(type) {
	case *net.TCPAddr:
		return addr.IP, addr.Port
	case *net.UDPAddr:
		return addr.IP, addr.Port
	}

	return nil, 0
}

func IPv4bcast() (ip net.IP) { return net.IP{255, 255, 255, 255} }

func IPv4allsys() (ip net.IP) { return net.IP{224, 0, 0, 1} }

func IPv4allrouter() (ip net.IP) { return net.IP{224, 0, 0, 2} }

func IPv4Zero() (ip net.IP) { return net.IP{0, 0, 0, 0} }

func IPv6Zero() (ip net.IP) {
	return net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
}

func ParseIP(s string) (ip net.IP, err error) {
	ip = net.ParseIP(s)
	if ip != nil {
		return ip, nil
	}

	return nil, fmt.Errorf("can not parse IP address (IP address=%v)", s)
}

func ParseIPv4(s string) (ip net.IP, err error) {
	ip, err = ParseIP(s)
	if err != nil {
		return nil, err
	}

	if ip = ip.To4(); ip == nil {
		return nil, fmt.Errorf("can not parse IP address (IPv4 address=%v)", s)
	}

	return ip, nil
}

func CloneIPNet(n *net.IPNet) (clone *net.IPNet) {
	if n == nil {
		return nil
	}

	return &net.IPNet{
		IP:   sliceutils.Clone(n.IP),
		Mask: net.IPMask(sliceutils.Clone(net.IP(n.Mask))),
	}
}

func ParseSubnet(s string) (n *net.IPNet, err error) {
	var ip net.IP

	// Detect if this is a CIDR or an IP early, so that the path to returning an
	// error is shorter.
	if !strings.Contains(s, "/") {
		ip, err = ParseIP(s)
		if err != nil {
			return nil, err
		}

		return SingleIPSubnet(ip), nil
	}

	ip, n, err = net.ParseCIDR(s)
	if err != nil {
		// Don't include the original error here, because it is basically
		// the same as ours but worse and has no additional information.
		return nil, err
	}

	if ip4 := ip.To4(); ip4 != nil {
		// Reduce the length of IP and mask if possible so that
		// IPNet.Contains doesn't waste time converting between 16- and
		// 4-byte versions.
		ip = ip4

		if ones, bits := n.Mask.Size(); ones >= 96 && bits == IPv6BitLen {
			// Copy the IPv4-length tail of the underlying slice to it's
			// beginning to avoid allocations in case of subsequent appending.
			copy(n.Mask, n.Mask[net.IPv6len-net.IPv4len:])
			n.Mask = n.Mask[:net.IPv4len]
		}
	}

	n.IP = ip

	return n, nil
}

func SingleIPSubnet(ip net.IP) (n *net.IPNet) {
	if ip4 := ip.To4(); ip4 != nil {
		return &net.IPNet{
			IP:   ip4,
			Mask: net.CIDRMask(IPv4BitLen, IPv4BitLen),
		}
	} else if len(ip) == net.IPv6len {
		return &net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(IPv6BitLen, IPv6BitLen),
		}
	}

	return nil
}

func ParseSubnets(ss ...string) (ns []*net.IPNet, err error) {
	l := len(ss)
	if l == 0 {
		return nil, nil
	}

	ns = make([]*net.IPNet, l)
	for i, s := range ss {
		ns[i], err = ParseSubnet(s)
		if err != nil {
			return nil, err
		}
	}

	return ns, nil
}

func ValidateIP(ip net.IP) (err error) {
	switch l := len(ip); l {
	case 0:
		return errors.New("invalid IP address for zero length string")
	case net.IPv4len, net.IPv6len:
		return nil
	default:
		return fmt.Errorf("invalid IP address length (length=%v)", l)
	}
}
