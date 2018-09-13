package netaddr

import (
	"fmt"
	"math/big"
	"net"
	"sort"

	"github.com/davecgh/go-spew/spew"
)

var scs = spew.ConfigState{Indent: "\t"}

type IPNetwork struct {
	start   *IPNumber
	version *Version
	Mask    *IPMask

	iteratorIndex int
}

func (nw *IPNetwork) String() string {
	ones, _ := nw.Mask.Size()
	return fmt.Sprintf("%s/%d", nw.start.ToIPAddress(), ones)
}

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

func (nw *IPNetwork) First() *IPAddress {
	return nw.start.ToIPAddress()
}
func (nw *IPNetwork) Last() *IPAddress {
	return nw.start.
		Add(nw.Length()).
		Sub(NewIPNumber(1)).
		ToIPAddress()
}

type IPMask struct {
	*net.IPMask
}

func (m *IPMask) Equals(other *IPMask) bool {
	maskInt := big.NewInt(0).SetBytes(*m.IPMask)
	otherInt := big.NewInt(0).SetBytes(*other.IPMask)
	return maskInt.Cmp(otherInt) == 0
}

func (m *IPMask) LessThan(other *IPMask) bool {
	maskInt := big.NewInt(0).SetBytes(*m.IPMask)
	otherInt := big.NewInt(0).SetBytes(*other.IPMask)
	return maskInt.Cmp(otherInt) == -1
}

func (m *IPMask) Int() {}

func MergeCIDRs(cidrs []IPNetwork) IPSet {
	var (
		merged IPSet
		ranges []IPRange
	)

	for _, cidr := range cidrs {
		ranges = append(ranges, IPRange {
			version: cidr.version,
			first: cidr.First(),
			last: cidr.Last(),
			network: &cidr})
	}

	sort.Sort(ByIPRanges(ranges))

	for i := len(ranges) - 1; i >= 0; i-- {
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

type Partition struct {
	Before    []*IPNetwork
	Partition *IPNetwork
	After     []*IPNetwork
}

func (nw *IPNetwork) Partition(exclude *IPNetwork) *Partition {

	scs.Dump(nw.Mask.Size())

	if exclude.Last().LessThan(nw.First()) {
		// Exclude subnet's upper bound address less than target
		// subnet's lower bound.
		fmt.Println("return1")
		return &Partition{
			After: []*IPNetwork{nw},
		}
	} else if nw.Last().LessThan(exclude.First()) {
		// Exclude subnet's lower bound address greater than target
		// subnet's upper bound.
		fmt.Println("return2")
		return &Partition{
			Before: []*IPNetwork{nw},
		}
	}

	scs.Dump(nw.PrefixLength())
	scs.Dump(exclude.PrefixLength())
	if nw.PrefixLength().GreaterThanOrEqual(exclude.PrefixLength()) {
		fmt.Println("return3")
		return &Partition{
			Partition: nw,
		}
	}

	var left []*IPNetwork
	var right []*IPNetwork

	targetModuleWidth := nw.version.bitLength
	newPrefixLength := nw.PrefixLength().Add(NewIPNumber(1))

	scs.Dump(targetModuleWidth)
	scs.Dump(newPrefixLength)

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
		fmt.Printf("exclude prefix length: %+v\n", exclude)
		fmt.Printf("newprefixlenght: %s\n", newPrefixLength)
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
	fmt.Println("return4")
	reverse(&right)
	return &Partition{
		Before:    left,
		Partition: exclude,
		After:     right,
	}
}

func reverse(slice *[]*IPNetwork) {
	s := *slice

	for i := 0; i < len(s)/2; i++ {
		j := len(s) - 1 - i
		s[i], s[j] = s[j], s[i]
	}
}

func (nw *IPNetwork) PrefixLength() *IPNumber {
	ones, _ := nw.Mask.Size()
	return NewIPNumber(int64(ones))
}

// NewNetworkFromInt returns a new network from an ipaddress integer with the default mask of all ones.
func newNetworkFromIP(version *Version, value *IPAddress) *IPNetwork {
	mask := net.CIDRMask(int(version.bitLength), int(version.bitLength))
	return &IPNetwork{
		start:   value.ToInt(),
		version: version,
		Mask:    &IPMask{IPMask: &mask},
	}
}

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
// IPAddresses are represented here as iPNetworks with a mask of /32
type IPSet []*IPNetwork

// Remote Removes an IP address or subnet or IPRange from this IP set. Does
// nothing if it is not already a member.
func (set *IPSet) Remove() {}

// Add adds an IPAddress or IPNetwork to this IPSet.
//
// IPAddresses are represented as IPNetworks with a /32 subnet mask, and where
// possible the IPAddresses and IPNetworks are merged with other members of the
// set to form more concise CIDR blocks.
func (set *IPSet) Add() {}

// Pop removes an arbitrary subnet from this IPSet
func (set *IPSet) Pop() {}

func (nw *IPNetwork) ContainsAddress(addr *IPAddress) bool {
	return nw.First().LessThan(addr) && addr.LessThan(nw.Last())
}

func (nw *IPNetwork) ContainsSubnetwork(other *IPNetwork) bool {
	return nw.First().LessThan(other.First()) &&
		nw.Last().GreaterThan(other.Last())
}

// returns the number of valid ip addresses in a subnet
func (m *IPMask) Length() *IPNumber {
	ones, bits := net.IPMask(*m.IPMask).Size()
	return NewIPNumber(2).Exp(NewIPNumber(int64(bits - ones)))
}

func (nw *IPNetwork) Length() *IPNumber { return nw.Mask.Length() }

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
