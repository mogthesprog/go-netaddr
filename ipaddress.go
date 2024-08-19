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
	// ErrorAddressOutOFBounds is an error returned when an IP number exceeds the IP version boundary.
	ErrorAddressOutOFBounds = fmt.Errorf("ip number out range of ip-version boundary")
)

var (
	// IPv4 represents the properties of IPv4, including its bit length and maximum address.
	IPv4 = &Version{
		number:    4,
		length:    4,
		bitLength: IPv4len * 8,
		max: &IPNumber{
			Int: (&big.Int{}).Sub((&big.Int{}).Exp(big.NewInt(2), big.NewInt(32), nil), big.NewInt(1)),
		},
	}

	// IPv6 represents the properties of IPv6, including its bit length and maximum address.
	IPv6 = &Version{
		number:    6,
		length:    16,
		bitLength: IPv6len * 8,
		max: &IPNumber{
			Int: (&big.Int{}).Sub((&big.Int{}).Exp(big.NewInt(2), big.NewInt(128), nil), big.NewInt(1)),
		},
	}
)

type (
	// IPNumber is the integer representation of an IP address.
	IPNumber struct{ *big.Int }

	// IPAddress represents an IP address and its version (IPv4 or IPv6).
	IPAddress struct {
		*net.IP
		version *Version
	}

	// Version represents the IP version (IPv4 or IPv6) and its properties.
	Version struct {
		number    int64
		length    int64
		bitLength int64
		max       *IPNumber
	}
)

// LessThan compares two IP Address versions, v and other. Returns true if v is less than other.
func (v *Version) LessThan(other *Version) bool {
	return v.length < other.length
}

// NewMask returns a new IPMask object with the passed ones and bits.
//
// Example usage:
//
//	mask := netaddr.NewMask(24, 32)
//	fmt.Println(mask)
func NewMask(ones, bits int64) *IPMask {
	mask := net.CIDRMask(int(ones), int(bits))
	return &IPMask{
		IPMask: &mask,
	}
}

// NewIP returns a new IPAddress object, initialized with the IP info parsed from ip.
//
// Example usage:
//
//	ip := netaddr.NewIP("192.168.1.1")
//	fmt.Println(ip)
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

// NewIPNumber returns an IPNumber for the passed number.
//
// Example usage:
//
//	ipNum := netaddr.NewIPNumber(123456)
//	fmt.Println(ipNum)
func NewIPNumber(v int64) *IPNumber {
	return &IPNumber{
		Int: big.NewInt(v),
	}
}

// String returns the string representation of version v.
//
// Example usage:
//
//	fmt.Println(netaddr.IPv4.String()) // Output: "IPv4"
func (v *Version) String() string {
	if v == IPv4 {
		return "IPv4"
	} else if v == IPv6 {
		return "IPv6"
	} else {
		return ""
	}
}

// String returns the string representation of address ip.
//
// Example usage:
//
//	ip := netaddr.NewIP("192.168.1.1")
//	fmt.Println(ip.String()) // Output: "192.168.1.1"
func (ip *IPAddress) String() string {
	return ip.IP.String()
}

// Version returns the IP version for IPAddress, ip.
//
// Example usage:
//
//	ip := netaddr.NewIP("192.168.1.1")
//	fmt.Println(ip.Version().String()) // Output: "IPv4"
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
//
// Example usage:
//
//	ip := netaddr.NewIP("192.168.1.1")
//	val := netaddr.NewIPNumber(1)
//	ip, err := ip.Increment(val)
//	if err != nil {
//	    fmt.Println(err)
//	}
//	fmt.Println(ip) // Output: "192.168.1.2"
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

// ValidIPV4 returns true when the passed bytes are a valid IPV4.
//
// Example usage:
//
//	ip := net.ParseIP("192.168.1.1")
//	fmt.Println(netaddr.ValidIPV4(ip)) // Output: true
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

// ToInt returns the integer representation (IPNumber) for the given IPAddress.
//
// Example usage:
//
//	ip := netaddr.NewIP("192.168.1.1")
//	ipNum := ip.ToInt()
//	fmt.Println(ipNum)
func (ip *IPAddress) ToInt() *IPNumber {
	num := NewIPNumber(0)
	num.SetBytes(*ip.IP)
	return num
}

// ToIPAddress converts the given IPNumber object to an IPAddress.
//
// Example usage:
//
//	ipNum := netaddr.NewIPNumber(3232235777)
//	ip := ipNum.ToIPAddress()
//	fmt.Println(ip.String()) // Output: "192.168.1.1"
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

// GreaterThan compares two IPNumbers, returning true when num is greater than other.
//
// Example usage:
//
//	ipNum1 := netaddr.NewIPNumber(3232235777) // 192.168.1.1
//	ipNum2 := netaddr.NewIPNumber(3232235778) // 192.168.1.2
//	fmt.Println(ipNum1.GreaterThan(ipNum2)) // Output: false
func (num *IPNumber) GreaterThan(other *IPNumber) bool {
	cmp := num.Cmp(other.Int)
	return cmp == 1
}

// GreaterThanOrEqual compares two IPNumbers, returning true when num is greater than or equal to other.
//
// Example usage:
//
//	ipNum1 := netaddr.NewIPNumber(3232235777) // 192.168.1.1
//	ipNum2 := netaddr.NewIPNumber(3232235778) // 192.168.1.2
//	fmt.Println(ipNum1.GreaterThanOrEqual(ipNum2)) // Output: false
func (num *IPNumber) GreaterThanOrEqual(other *IPNumber) bool {
	if cmp := num.Cmp(other.Int); cmp >= 0 {
		return true
	}
	return false
}

// LessThan compares two IPNumbers, returning true when num is less than other.
//
// Example usage:
//
//	ipNum1 := netaddr.NewIPNumber(3232235777) // 192.168.1.1
//	ipNum2 := netaddr.NewIPNumber(3232235778) // 192.168.1.2
//	fmt.Println(ipNum1.LessThan(ipNum2)) // Output: true
func (num *IPNumber) LessThan(other *IPNumber) bool {
	cmp := num.Cmp(other.Int)
	return cmp == -1
}

// LessThanOrEqual compares two IPNumbers, returning true when num is less than or equal to other.
//
// Example usage:
//
//	ipNum1 := netaddr.NewIPNumber(3232235777) // 192.168.1.1
//	ipNum2 := netaddr.NewIPNumber(3232235778) // 192.168.1.2
//	fmt.Println(ipNum1.LessThanOrEqual(ipNum2)) // Output: true
func (num *IPNumber) LessThanOrEqual(other *IPNumber) bool {
	if cmp := num.Cmp(other.Int); cmp <= 0 {
		return true
	}
	return false
}

// Equal compares two IPNumbers, returning true when num is equal to other.
//
// Example usage:
//
//	ipNum1 := netaddr.NewIPNumber(3232235777) // 192.168.1.1
//	ipNum2 := netaddr.NewIPNumber(3232235777) // 192.168.1.1
//	fmt.Println(ipNum1.Equal(ipNum2)) // Output: true
func (num *IPNumber) Equal(other *IPNumber) bool {
	cmp := num.Cmp(other.Int)
	return cmp == 0
}

// Add adds two IPNumbers and returns the result.
//
// Example usage:
//
//	ipNum1 := netaddr.NewIPNumber(3232235777) // 192.168.1.1
//	ipNum2 := netaddr.NewIPNumber(1)
//	result := ipNum1.Add(ipNum2)
//	fmt.Println(result) // Output: 3232235778
func (num *IPNumber) Add(v *IPNumber) *IPNumber {
	int := big.NewInt(0).Add(num.Int, v.Int)
	return &IPNumber{int}
}

// Sub subtracts v from num and returns the result.
//
// Example usage:
//
//	ipNum1 := netaddr.NewIPNumber(3232235778) // 192.168.1.2
//	ipNum2 := netaddr.NewIPNumber(1)
//	result := ipNum1.Sub(ipNum2)
//	fmt.Println(result) // Output: 3232235777
func (num *IPNumber) Sub(v *IPNumber) *IPNumber {
	int := big.NewInt(0).Sub(num.Int, v.Int)
	return &IPNumber{int}
}

// Exp raises num to the power of v and returns the result.
//
// Example usage:
//
//	ipNum := netaddr.NewIPNumber(2)
//	exp := netaddr.NewIPNumber(8)
//	result := ipNum.Exp(exp)
//	fmt.Println(result) // Output: 256
func (num *IPNumber) Exp(v *IPNumber) *IPNumber {
	int := big.NewInt(0).Exp(num.Int, v.Int, nil)
	return &IPNumber{int}
}

// And performs a bitwise AND operation on num and v, returning the result.
//
// Example usage:
//
//	ipNum1 := netaddr.NewIPNumber(3232235777) // 192.168.1.1
//	ipNum2 := netaddr.NewIPNumber(255)
//	result := ipNum1.And(ipNum2)
//	fmt.Println(result) // Output: 1
func (num *IPNumber) And(v *IPNumber) *IPNumber {
	int := big.NewInt(0).And(num.Int, v.Int)
	return &IPNumber{int}
}

// Lsh shifts num left by v bits and returns the result.
//
// Example usage:
//
//	ipNum := netaddr.NewIPNumber(1)
//	result := ipNum.Lsh(8)
//	fmt.Println(result) // Output: 256
func (num *IPNumber) Lsh(v uint) *IPNumber {
	int := big.NewInt(0).Lsh(num.Int, v)
	return &IPNumber{int}
}

// Neg returns the negative of num.
//
// Example usage:
//
//	ipNum := netaddr.NewIPNumber(1)
//	result := ipNum.Neg()
//	fmt.Println(result) // Output: -1
func (num *IPNumber) Neg() *IPNumber {
	int := big.NewInt(0).Neg(num.Int)
	return &IPNumber{int}
}

// MinAddress returns the smaller of two IP addresses.
//
// Example usage:
//
//	addr1 := netaddr.NewIP("192.168.1.1")
//	addr2 := netaddr.NewIP("192.168.1.2")
//	fmt.Println(netaddr.MinAddress(addr1, addr2)) // Output: "192.168.1.1"
func MinAddress(addr1, addr2 *IPAddress) *IPAddress {
	if addr1.ToInt().LessThanOrEqual(addr2.ToInt()) {
		return addr1
	}
	return addr2
}

// LessThan compares two IPAddresses, returning true when ip is less than other.
//
// Example usage:
//
//	ip1 := netaddr.NewIP("192.168.1.1")
//	ip2 := netaddr.NewIP("192.168.1.2")
//	fmt.Println(ip1.LessThan(ip2)) // Output: true
func (ip *IPAddress) LessThan(other *IPAddress) bool {
	return ip.ToInt().LessThan(other.ToInt())
}

// GreaterThan compares two IPAddresses, returning true when ip is greater than other.
//
// Example usage:
//
//	ip1 := netaddr.NewIP("192.168.1.2")
//	ip2 := netaddr.NewIP("192.168.1.1")
//	fmt.Println(ip1.GreaterThan(ip2)) // Output: true
func (ip *IPAddress) GreaterThan(other *IPAddress) bool {
	return ip.ToInt().GreaterThan(other.ToInt())
}

// LessThanOrEqual compares two IPAddresses, returning true when ip is less than or equal to other.
//
// Example usage:
//
//	ip1 := netaddr.NewIP("192.168.1.1")
//	ip2 := netaddr.NewIP("192.168.1.2")
//	fmt.Println(ip1.LessThanOrEqual(ip2)) // Output: true
func (ip *IPAddress) LessThanOrEqual(other *IPAddress) bool {
	return ip.ToInt().LessThanOrEqual(other.ToInt())
}

// Equal compares two IPAddresses, returning true when ip is equal to other.
//
// Example usage:
//
//	ip1 := netaddr.NewIP("192.168.1.1")
//	ip2 := netaddr.NewIP("192.168.1.1")
//	fmt.Println(ip1.Equal(ip2)) // Output: true
func (ip *IPAddress) Equal(other *IPAddress) bool {
	return ip.ToInt().Equal(other.ToInt())
}

// GreaterThanOrEqual compares two IPAddresses, returning true when ip is greater than or equal to other.
//
// Example usage:
//
//	ip1 := netaddr.NewIP("192.168.1.2")
//	ip2 := netaddr.NewIP("192.168.1.1")
//	fmt.Println(ip1.GreaterThanOrEqual(ip2)) // Output: true
func (ip *IPAddress) GreaterThanOrEqual(other *IPAddress) bool {
	return ip.ToInt().GreaterThanOrEqual(other.ToInt())
}
