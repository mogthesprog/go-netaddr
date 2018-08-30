package netaddr

import (
	"fmt"
	"math/big"
	"net"
)

const (
	// IP address lengths (bytes).
	IPv4len = 4
	IPv6len = 16
)

var (
	ErrorAddressOutOFBounds = fmt.Errorf("ip number out range of ip-version boundary")
)

var (
	IPv4 = &Version{
		number:    4,
		length:    4,
		bitLength: IPv4len * 8,
		// IPv4 max is 2**32, represented as big.Int
		max: &IPNumber{
			Int: (&big.Int{}).Sub((&big.Int{}).Exp(big.NewInt(2), big.NewInt(32), nil), big.NewInt(1)),
		},
	}

	IPv6 = &Version{
		number:    6,
		length:    16,
		bitLength: IPv6len * 8,
		// IPv4 max is 2**128, represented as big.Int
		max: &IPNumber{
			Int: (&big.Int{}).Sub((&big.Int{}).Exp(big.NewInt(2), big.NewInt(128), nil), big.NewInt(1)),
		},
	}
)

type (
	// IPNumber is the integer representation of an IP address
	IPNumber  struct{ *big.Int }
	IPAddress struct {
		*net.IP
		version *Version
	}

	Version struct {
		number    int64
		length    int64
		bitLength int64
		max       *IPNumber
	}
)

func (v *Version) LessThan(other *Version) bool {
	return v.length < other.length
}

func NewMask(ones, bits int64) *IPMask {
	mask := net.CIDRMask(int(ones), int(bits))
	return &IPMask{
		IPMask: &mask,
	}
}

func NewIP(ip string) *IPAddress {
	newIP := net.ParseIP(ip)
	if newIP.To4() != nil {
		newIP = newIP.To4()
		return &IPAddress{
			IP:      &newIP,
			version: IPv4,
		}
	}

	newIP = newIP.To16()
	return &IPAddress{
		IP:      &newIP,
		version: IPv6,
	}
}

func NewIPNumber(v int64) *IPNumber {
	return &IPNumber{
		Int: big.NewInt(v),
	}
}

func (v *Version) String() string {
	if v == IPv4 {
		return "IPv4"
	} else if v == IPv6 {
		return "IPv6"
	} else {
		return ""
	}
}

func (ip *IPAddress) String() string {
	return fmt.Sprintf("Version: %s , IP: %s", ip.version, ip.IP)
}

func (ip *IPAddress) Version() *Version {
	if len(*ip.IP) == IPv6len {
		return IPv6
	}
	if len(*ip.IP) == IPv4len {
		return IPv4
	}
	return nil
}

// Increment increments the IPAddress by an amount, val, which is of big.Int type.
func (ip *IPAddress) Increment(val *IPNumber) (*IPAddress, error) {
	ipNum := ip.ToInt()
	if ipNum.Equal(NewIPNumber(0)) {
		return ip, nil
	}
	ipNum = ipNum.Add(val)
	if ipNum.GreaterThanOrEqual(NewIPNumber(0)) &&
		ipNum.LessThanOrEqual(ip.Version().max) {
		ip.IP = ipNum.ToIPAddress().IP
		return ip, nil
	}

	return nil, ErrorAddressOutOFBounds
}

func ValidIPV4(ipBytes []byte) bool {
	// we need to check for length 0, since big.Int has length 0 for big.Int(0)
	if len(ipBytes) == IPv4len || len(ipBytes) == 0 {
		return true
	} else if len(ipBytes) != IPv6len {
		return false
	}

	ipv4Flag := true
	// Check if the first 12 bytes are all zero (i.e. IPv4)
	for i, v := range ipBytes {
		// 11 is the last index of ipv6 ONLY address bytes
		if v == 0 && i <= IPv6len-IPv4len-1 {
			ipv4Flag = false
		}
	}

	return ipv4Flag
}

func (ip *IPAddress) ToInt() *IPNumber {
	num := NewIPNumber(0)
	num.SetBytes(*ip.IP)
	return num
}

func (num *IPNumber) ToIPAddress() *IPAddress {
	var (
		bytes   net.IP
		version *Version
	)
	// get the bytes of bigInt
	bigintBytes := num.Bytes()

	if ValidIPV4(bigintBytes) {
		bytes = make(net.IP, 4)
		version = IPv4
	} else {
		bytes = make(net.IP, 16)
		version = IPv6
	}

	for i := 0; i < len(bytes); i++ {
		// Handle the case where len(bigintbytes) == 0. This is the case for a
		// zero big.Int type.
		if len(bigintBytes) == i {
			break
		}
		bytes[len(bytes)-(i+1)] = bigintBytes[len(bigintBytes)-(i+1)]
	}
	return &IPAddress{
		IP:      &bytes,
		version: version,
	}
}

func (num *IPNumber) GreaterThan(other *IPNumber) bool {
	cmp := num.Cmp(other.Int)
	if cmp == 1 {
		return true
	}
	return false
}
func (num *IPNumber) GreaterThanOrEqual(other *IPNumber) bool {
	if cmp := num.Cmp(other.Int); cmp >= 0 {
		return true
	}
	return false
}

func (ip *IPAddress) LessThan(other *IPAddress) bool    { return ip.ToInt().LessThan(other.ToInt()) }
func (ip *IPAddress) GreaterThan(other *IPAddress) bool { return ip.ToInt().GreaterThan(other.ToInt()) }
func (ip *IPAddress) LessThanOrEqual(other *IPAddress) bool {
	return ip.ToInt().LessThanOrEqual(other.ToInt())
}
func (ip *IPAddress) Equal(other *IPAddress) bool { return ip.ToInt().Equal(other.ToInt()) }
func (ip *IPAddress) GreaterThanOrEqual(other *IPAddress) bool {
	return ip.ToInt().GreaterThanOrEqual(other.ToInt())
}

func (num *IPNumber) LessThan(other *IPNumber) bool {
	cmp := num.Cmp(other.Int)
	if cmp == -1 {
		return true
	}
	return false
}
func (num *IPNumber) LessThanOrEqual(other *IPNumber) bool {
	if cmp := num.Cmp(other.Int); cmp <= 0 {
		return true
	}
	return false
}
func (num *IPNumber) Equal(other *IPNumber) bool {
	cmp := num.Cmp(other.Int)
	if cmp == 0 {
		return true
	}
	return false

}
func MinAddress(addr1, addr2 *IPAddress) *IPAddress {
	if addr1.ToInt().LessThanOrEqual(addr2.ToInt()) {
		return addr1
	}
	return addr2
}

func (num *IPNumber) Add(v *IPNumber) *IPNumber {
	int := big.NewInt(0).Add(num.Int, v.Int)
	return &IPNumber{int}
}

func (num *IPNumber) Sub(v *IPNumber) *IPNumber {
	int := big.NewInt(0).Sub(num.Int, v.Int)
	return &IPNumber{int}
}

func (num *IPNumber) Exp(v *IPNumber) *IPNumber {
	int := big.NewInt(0).Exp(num.Int, v.Int, nil)
	return &IPNumber{int}
}

func (num *IPNumber) And(v *IPNumber) *IPNumber {
	int := big.NewInt(0).And(num.Int, v.Int)
	return &IPNumber{int}
}

func (num *IPNumber) Lsh(v uint) *IPNumber {
	int := big.NewInt(0).Lsh(num.Int, v)
	return &IPNumber{int}
}

func (num *IPNumber) Neg() *IPNumber {
	int := big.NewInt(0).Neg(num.Int)
	return &IPNumber{int}
}
