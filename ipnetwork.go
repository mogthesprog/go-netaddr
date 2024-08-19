package netaddr

import (
	"fmt"
	"math"
	"math/big"
	"net"
	"sort"
)

// IPNetwork defines an IPAddress network, including version and mask.
type IPNetwork struct {
	start   *IPNumber
	version *Version
	Mask    *IPMask
}

// String returns the string representation of the network, e.g., "127.0.0.1/8".
//
// Example usage:
//
//	nw, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	fmt.Println(nw.String()) // Output: "192.168.1.0/24"
func (nw *IPNetwork) String() string {
	ones, _ := nw.Mask.Size()
	return fmt.Sprintf("%s/%d", nw.start.ToIPAddress(), ones)
}

// NewIPNetwork creates a new IPNetwork from a CIDR string.
//
// Example usage:
//
//	nw, err := netaddr.NewIPNetwork("192.168.1.0/24")
//	if err != nil {
//	    fmt.Println(err)
//	}
//	fmt.Println(nw)
func NewIPNetwork(cidr string) (*IPNetwork, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	version := &Version{}
	_, width := network.Mask.Size()
	if width == IPv6len*8 {
		version = IPv6
	}
	if width == IPv4len*8 {
		version = IPv4
	}

	addr := &IPAddress{IP: &network.IP}
	return &IPNetwork{
		start:   addr.ToInt(),
		version: version,
		Mask:    &IPMask{IPMask: &network.Mask},
	}, nil
}

// newNetworkFromBoundaries creates a new IPNetwork from two IP addresses
// representing the first and last addresses in the network.
//
// Example usage:
//
//	first := netaddr.NewIP("192.168.1.0")
//	last := netaddr.NewIP("192.168.1.255")
//	network, err := netaddr.newNetworkFromBoundaries(first, last)
//	if err != nil {
//	    fmt.Println(err)
//	}
//	fmt.Println(network)
func newNetworkFromBoundaries(first, last *IPAddress) (*IPNetwork, error) {
	if first.Version() != last.Version() {
		return nil, fmt.Errorf("version of input addresses, first: %d, last: %d, don't match", first.Version().number, last.Version().number)
	}

	ipNumber := last.ToInt()
	lowestIPNumber := first.ToInt()
	width := first.version.bitLength
	prefixlen := NewIPNumber(width)

	// Search outwards from the longest prefix to find the correct prefix
	for prefixlen.GreaterThan(NewIPNumber(0)) && ipNumber.GreaterThan(lowestIPNumber) {
		prefixlen = prefixlen.Sub(NewIPNumber(1))
		// The below represents ipNum &= -(1 << (uint64(width) - uint64(prefixlen)))
		ipNumber = ipNumber.And(NewIPNumber(1).Lsh(uint(width - prefixlen.Int64())).Neg())
	}

	mask := NewMask(prefixlen.Int64(), width)

	return &IPNetwork{
		start:   ipNumber,
		version: first.Version(),
		Mask:    mask,
	}, nil
}

// First returns the first IP address in the network.
//
// Example usage:
//
//	nw, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	first := nw.First()
//	fmt.Println(first) // Output: "192.168.1.0"
func (nw *IPNetwork) First() *IPAddress {
	return nw.start.ToIPAddress()
}

// Last returns the last IP address in the network.
//
// Example usage:
//
//	nw, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	last := nw.Last()
//	fmt.Println(last) // Output: "192.168.1.255"
func (nw *IPNetwork) Last() *IPAddress {
	return nw.start.
		Add(nw.Length()).
		Sub(NewIPNumber(1)).
		ToIPAddress()
}

// IPMask represents a subnet mask.
type IPMask struct {
	*net.IPMask
}

// Equals compares two IPMasks and returns true if they are equal.
//
// Example usage:
//
//	mask1 := netaddr.NewMask(24, 32)
//	mask2 := netaddr.NewMask(24, 32)
//	fmt.Println(mask1.Equals(mask2)) // Output: true
func (m *IPMask) Equals(other *IPMask) bool {
	maskInt := big.NewInt(0).SetBytes(*m.IPMask)
	otherInt := big.NewInt(0).SetBytes(*other.IPMask)
	return maskInt.Cmp(otherInt) == 0
}

// LessThan compares two IPMasks and returns true if the mask is less than the other.
//
// Example usage:
//
//	mask1 := netaddr.NewMask(24, 32)
//	mask2 := netaddr.NewMask(16, 32)
//	fmt.Println(mask1.LessThan(mask2)) // Output: false
func (m *IPMask) LessThan(other *IPMask) bool {
	maskInt := big.NewInt(0).SetBytes(*m.IPMask)
	otherInt := big.NewInt(0).SetBytes(*other.IPMask)
	return maskInt.Cmp(otherInt) == -1
}

// MergeCIDRs merges a slice of IPNetwork objects into an IPSet.
//
// Example usage:
//
//	cidr1, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	cidr2, _ := netaddr.NewIPNetwork("192.168.2.0/24")
//	merged := netaddr.MergeCIDRs([]netaddr.IPNetwork{*cidr1, *cidr2})
//	fmt.Println(merged)
func MergeCIDRs(cidrs []IPNetwork) IPSet {
	var (
		merged IPSet
		ranges []IPRange
	)

	for _, cidr := range cidrs {
		ranges = append(ranges, IPRange{
			version: cidr.version,
			first:   cidr.First(),
			last:    cidr.Last(),
			network: &cidr})
	}

	sort.Sort(ByIPRanges(ranges))

	for i := len(ranges) - 1; i > 0; i-- {
		current := ranges[i]
		next := ranges[i-1]
		if current.version == next.version &&
			current.first.ToInt().LessThan(next.last.ToInt()) {
			ranges[i-1] = struct {
				version *Version
				first   *IPAddress
				last    *IPAddress
				network *IPNetwork
			}{version: current.version, first: current.last, last: MinAddress(next.first, current.first), network: &IPNetwork{}}
		}
	}

	for _, value := range ranges {
		if value.network == nil {
			merged = append(merged, value.network)
		} else {
			subnets, err := IPRangeToCIDRS(value.version, value.first, value.last)
			if err != nil {
				// do something
			}
			merged = append(merged, subnets...)
		}
	}

	return merged
}

// Partition defines a structure to hold the parts of an IP network before, during, and after partitioning.
type Partition struct {
	Before    []*IPNetwork
	Partition *IPNetwork
	After     []*IPNetwork
}

// Partition divides the IPNetwork into three parts: the portion before the exclude network,
// the partition that overlaps with the exclude network, and the portion after.
//
// Example usage:
//
//	nw, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	exclude, _ := netaddr.NewIPNetwork("192.168.1.128/25")
//	partition := nw.Partition(exclude)
//	fmt.Println(partition)
func (nw *IPNetwork) Partition(exclude *IPNetwork) *Partition {

	if exclude.Last().LessThan(nw.First()) {
		// Exclude subnet's upper bound address less than target
		// subnet's lower bound.
		return &Partition{
			After: []*IPNetwork{nw},
		}
	} else if nw.Last().LessThan(exclude.First()) {
		// Exclude subnet's lower bound address greater than target
		// subnet's upper bound.
		return &Partition{
			Before: []*IPNetwork{nw},
		}
	}

	if nw.PrefixLength().GreaterThanOrEqual(exclude.PrefixLength()) {
		return &Partition{
			Partition: nw,
		}
	}

	var left []*IPNetwork
	var right []*IPNetwork

	targetModuleWidth := nw.version.bitLength
	newPrefixLength := nw.PrefixLength().Add(NewIPNumber(1))

	targetFirst := nw.First().ToInt()
	version := exclude.version
	iLower := targetFirst

	// Upper IP
	iUpper := targetFirst.Add(
		NewIPNumber(2).
			Exp(NewIPNumber(targetModuleWidth).
				Sub(newPrefixLength)),
	)

	for {
		if exclude.PrefixLength().LessThan(newPrefixLength) {
			break
		}
		var matched *IPNumber
		if exclude.First().ToInt().GreaterThanOrEqual(iUpper) {
			exclude := newNetworkFromIP(version, iLower.ToIPAddress())
			exclude.Mask = NewMask(newPrefixLength.Int64(), version.bitLength)
			left = append(left, exclude)
			matched = iUpper
		} else {
			exclude := newNetworkFromIP(version, iUpper.ToIPAddress())
			exclude.Mask = NewMask(newPrefixLength.Int64(), version.bitLength)
			right = append(right, exclude)
			matched = iLower
		}

		newPrefixLength = newPrefixLength.Add(NewIPNumber(1))

		if newPrefixLength.GreaterThan(NewIPNumber(targetModuleWidth)) {
			break
		}

		iLower = matched
		iUpper = matched.Add(
			NewIPNumber(2).
				Exp(NewIPNumber(targetModuleWidth).
					Sub(newPrefixLength)),
		)
	}
	reverse(&right)
	return &Partition{
		Before:    left,
		Partition: exclude,
		After:     right,
	}
}

// Subnet divides a network into smaller subnets based on the provided CIDR prefix.
//
// Example usage:
//
//	nw, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	subnets, err := nw.Subnet(25)
//	if err != nil {
//	    fmt.Println(err)
//	}
//	for _, subnet := range subnets {
//	    fmt.Println(subnet)
//	}
func (nw *IPNetwork) Subnet(newCIDRPrefix int) ([]*IPNetwork, error) {
	thisCidrPrefix, addressBits := nw.Mask.Size()
	if !(0 <= thisCidrPrefix || thisCidrPrefix <= addressBits) {
		return nil, fmt.Errorf("prefix %d is not valid", thisCidrPrefix)
	}

	if thisCidrPrefix > newCIDRPrefix {
		return []*IPNetwork{}, nil
	}
	maxNoSubnets := int(math.Pow(2, float64(addressBits-thisCidrPrefix)) / math.Pow(2, float64(addressBits-newCIDRPrefix)))
	var results []*IPNetwork
	for i := 0; i < maxNoSubnets; i++ {
		newCIDR := fmt.Sprintf("%s/%d", nw.First().IP, newCIDRPrefix)
		newSubnet, err := NewIPNetwork(newCIDR)
		if err != nil {
			return nil, err
		}
		sL := newSubnet.Length()
		sL.Mul(sL.Int, big.NewInt(int64(i)))
		newSubnet.start = newSubnet.start.Add(sL)
		results = append(results, newSubnet)
	}
	return results, nil
}

// reverse reverses the order of the slice of IPNetwork pointers.
//
// Example usage:
//
//	slice := []*netaddr.IPNetwork{nw1, nw2, nw3}
//	netaddr.reverse(&slice)
//	fmt.Println(slice)
func reverse(slice *[]*IPNetwork) {
	s := *slice

	for i := 0; i < len(s)/2; i++ {
		j := len(s) - 1 - i
		s[i], s[j] = s[j], s[i]
	}
}

// PrefixLength returns the prefix length of the IP network mask.
//
// Example usage:
//
//	nw, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	fmt.Println(nw.PrefixLength()) // Output: 24
func (nw *IPNetwork) PrefixLength() *IPNumber {
	ones, _ := nw.Mask.Size()
	return NewIPNumber(int64(ones))
}

// newNetworkFromIP returns a new network from an IP address with the default mask of all ones.
//
// Example usage:
//
//	ip := netaddr.NewIP("192.168.1.1")
//	network := netaddr.newNetworkFromIP(netaddr.IPv4, ip)
//	fmt.Println(network)
func newNetworkFromIP(version *Version, value *IPAddress) *IPNetwork {
	mask := net.CIDRMask(int(version.bitLength), int(version.bitLength))
	return &IPNetwork{
		start:   value.ToInt(),
		version: version,
		Mask:    &IPMask{IPMask: &mask},
	}
}

// IPRangeToCIDRS converts an IP range defined by a start and end address to a list of CIDR blocks.
//
// Example usage:
//
//	start := netaddr.NewIP("192.168.1.0")
//	end := netaddr.NewIP("192.168.1.255")
//	cidrs, err := netaddr.IPRangeToCIDRS(netaddr.IPv4, start, end)
//	if err != nil {
//	    fmt.Println(err)
//	}
//	for _, cidr := range cidrs {
//	    fmt.Println(cidr)
//	}
func IPRangeToCIDRS(version *Version, start, end *IPAddress) ([]*IPNetwork, error) {

	var cidrs []*IPNetwork

	subnet, err := newNetworkFromBoundaries(start, end)
	if err != nil {
		return nil, err
	}

	if subnet.First().LessThan(start) {
		excludeAddress := start
		_, err := excludeAddress.Increment(NewIPNumber(-1))
		if err != nil {
			return nil, err
		}
		exclude := newNetworkFromIP(version, excludeAddress)
		afterPartition := subnet.Partition(exclude).After
		cidrs = append(cidrs, afterPartition...)
		lastCidrIndex := len(cidrs) - 1
		if lastCidrIndex >= 0 {
			subnet = cidrs[lastCidrIndex]
			// Remove the last element of cidrs
			cidrs[lastCidrIndex] = &IPNetwork{}
			cidrs = cidrs[:lastCidrIndex]
		}
	}

	if subnet.Last().GreaterThan(end) {
		excludeAddress := end
		excludeAddress, err := excludeAddress.Increment(NewIPNumber(1))
		if err != nil && err != ErrorAddressOutOFBounds {
			return nil, err
		}
		exclude := newNetworkFromIP(version, excludeAddress)
		beforePartition := subnet.Partition(exclude).Before
		cidrs = append(cidrs, beforePartition...)
	} else {
		cidrs = append(cidrs, subnet)
	}

	return cidrs, nil
}

// IPSet represents an unordered collection of unique IP addresses and subnets.
// IPAddresses are represented here as IPNetworks with a mask of /32
type IPSet []*IPNetwork

// Remove removes an IP address or subnet from this IPSet. Does nothing if it is not already a member.
//
// Example usage:
//
//	set := netaddr.IPSet{nw1, nw2}
//	set.Remove(nw1)
//	fmt.Println(set)
func (set *IPSet) Remove() {}

// Add adds an IP address or IP network to this IPSet.
// IP addresses are represented as IPNetworks with a /32 subnet mask, and where possible,
// the IP addresses and IPNetworks are merged with other members of the set to form more concise CIDR blocks.
//
// Example usage:
//
//	set := netaddr.IPSet{}
//	set.Add(nw1)
//	fmt.Println(set)
func (set *IPSet) Add() {}

// Pop removes an arbitrary subnet from this IPSet.
//
// Example usage:
//
//	set := netaddr.IPSet{nw1, nw2}
//	set.Pop()
//	fmt.Println(set)
func (set *IPSet) Pop() {}

// ContainsAddress checks if the network contains a specific IP address.
//
// Example usage:
//
//	nw, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	ip := netaddr.NewIP("192.168.1.100")
//	fmt.Println(nw.ContainsAddress(ip)) // Output: true
func (nw *IPNetwork) ContainsAddress(addr *IPAddress) bool {
	return nw.First().LessThanOrEqual(addr) && addr.LessThanOrEqual(nw.Last())
}

// ContainsSubnetwork checks if the network contains another subnetwork.
//
// Example usage:
//
//	nw1, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	nw2, _ := netaddr.NewIPNetwork("192.168.1.128/25")
//	fmt.Println(nw1.ContainsSubnetwork(nw2)) // Output: true
func (nw *IPNetwork) ContainsSubnetwork(other *IPNetwork) bool {
	return nw.First().LessThanOrEqual(other.First()) &&
		nw.Last().GreaterThanOrEqual(other.Last())
}

// Length returns the number of valid IP addresses in a subnet.
//
// Example usage:
//
//	mask := netaddr.NewMask(24, 32)
//	fmt.Println(mask.Length()) // Output: 256
func (m *IPMask) Length() *IPNumber {
	ones, bits := net.IPMask(*m.IPMask).Size()
	return NewIPNumber(2).Exp(NewIPNumber(int64(bits - ones)))
}

// Length returns the number of valid IP addresses in the IPNetwork.
//
// Example usage:
//
//	nw, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	fmt.Println(nw.Length()) // Output: 256
func (nw *IPNetwork) Length() *IPNumber { return nw.Mask.Length() }

// Equal compares two IPNetworks for equality.
//
// Example usage:
//
//	nw1, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	nw2, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	fmt.Println(nw1.Equal(nw2)) // Output: true
func (nw *IPNetwork) Equal(other *IPNetwork) bool {
	if nw.version != other.version {
		return false
	}
	if !nw.Mask.Equals(other.Mask) {
		return false
	}
	if !nw.start.Equal(other.start) {
		return false
	}
	return true
}

// LessThan compares two IPNetworks, returning true if nw is less than other.
//
// Example usage:
//
//	nw1, _ := netaddr.NewIPNetwork("192.168.1.0/24")
//	nw2, _ := netaddr.NewIPNetwork("192.168.2.0/24")
//	fmt.Println(nw1.LessThan(nw2)) // Output: true
func (nw *IPNetwork) LessThan(other *IPNetwork) bool {
	if nw.version != other.version {
		return nw.version.LessThan(other.version)
	}
	if !nw.start.Equal(other.start) {
		return nw.start.LessThan(other.start)
	}
	if !nw.Mask.Equals(other.Mask) {
		return nw.Mask.LessThan(other.Mask)
	}
	return false
}
