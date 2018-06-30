package netaddr

import (
	"fmt"
	"net"

	"github.com/davecgh/go-spew/spew"
)

var scs = spew.ConfigState{Indent: "\t"}

type IPNetwork struct {
	start   *IPNumber
	version *Version
	Mask    *IPMask

	iteratorIndex int
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
		prefixlen.Sub(NewIPNumber(1))
		// The below represents ipNum &= -(1 << (uint64(width) - uint64(prefixlen)))
		ipNumber.And(NewIPNumber(1).Lsh(uint(width - prefixlen.Int64())).Neg())
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

func (mask *IPMask) Int() {}

func MergeCIDRs(cidrs []IPNetwork) IPSet {
	var (
		merged IPSet
		ranges []struct {
			version *Version
			first   *IPAddress
			last    *IPAddress
			network *IPNetwork
		}
	)

	for _, cidr := range cidrs {
		ranges = append(ranges, struct {
			version *Version
			first   *IPAddress
			last    *IPAddress
			network *IPNetwork
		}{version: cidr.version, first: cidr.First(), last: cidr.Last(), network: &cidr})
	}

	// TODO: Must sort ranges here

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
	before []*IPNetwork
	after  []*IPNetwork
}

func (nw *IPNetwork) Partition(exclude *IPNetwork) *Partition {
	fmt.Printf("Network: %+v\n\n", nw.First().IP.To4())
	fmt.Printf("exclude: %+v\n\n", exclude.First().IP.To4())
	if exclude.Last().LessThan(nw.First()) {
		// Exclude subnet's upper bound address less than target
		// subnet's lower bound.
		fmt.Println("return1")
		return &Partition{
			after: []*IPNetwork{nw},
		}
	} else if nw.Last().LessThan(exclude.First()) {
		// Exclude subnet's lower bound address greater than target
		// subnet's upper bound.
		fmt.Println("return2")
		return &Partition{
			before: []*IPNetwork{nw},
		}
	}

	if nw.PrefixLength().GreaterThanOrEqual(exclude.PrefixLength()) {
		fmt.Println("return3")
		return nil
	}

	var left []*IPNetwork
	var right []*IPNetwork

	targetModuleWidth := nw.version.length
	newPrefixLength := nw.PrefixLength().Add(NewIPNumber(1))

	subnetWidth := NewIPNumber(2).
		Exp(NewIPNumber(targetModuleWidth).Sub(newPrefixLength))

	targetFirst := nw.First().ToInt()
	version := exclude.version
	iLower := targetFirst

	// Upper IP
	iUpper := NewIPNumber(targetFirst.Int64()).Add(subnetWidth)

	for {
		var matched *IPNumber
		if exclude.First().ToInt().LessThanOrEqual(iUpper) {
			exclude := newNetworkFromIP(version, iLower.ToIPAddress())
			left = append(left, exclude)
			matched = iLower
		} else {
			exclude := newNetworkFromIP(version, iUpper.ToIPAddress())
			right = append(right, exclude)
			matched = iUpper
		}

		newPrefixLength.Add(NewIPNumber(1))

		if newPrefixLength.GreaterThan(NewIPNumber(targetModuleWidth)) {
			break
		}

		iLower = matched
		iUpper = matched.Add(subnetWidth)
	}

	fmt.Println("return4")
	return &Partition{
		before: left,
		after:  right,
	}
}

func (nw *IPNetwork) PrefixLength() *IPNumber {
	ones, _ := nw.Mask.Size()
	return NewIPNumber(int64(ones))
}

// NewNetworkFromInt returns a new network from an ipaddress integer with the default mask of all ones.
func newNetworkFromIP(version *Version, value *IPAddress) *IPNetwork {
	mask := net.CIDRMask(int(version.length), int(version.length))
	return &IPNetwork{
		start:   value.ToInt(),
		version: version,
		Mask:    &IPMask{IPMask: &mask},
	}
}

func IPRangeToCIDRS(version *Version, start, end *IPAddress) ([]*IPNetwork, error) {

	var (
		cidrs    []*IPNetwork
		cidrSpan []*IPNetwork
	)

	subnet, err := newNetworkFromBoundaries(start, end)
	if err != nil {
		return nil, err
	}

	scs.Dump(subnet.First())
	scs.Dump(start.ToInt())

	if subnet.First().LessThan(start) {
		excludeAddress := start
		fmt.Printf("excludeAddress: %s\n", excludeAddress)
		_, err := excludeAddress.Increment(NewIPNumber(-1))
		if err != nil {
			return nil, err
		}
		fmt.Printf("excludeAddress minux 1: %s\n", excludeAddress)
		exclude := newNetworkFromIP(version, excludeAddress)
		fmt.Printf("excludeAddress IP: %+v\n", exclude)
		cidrs = append(cidrs, subnet.Partition(exclude).after...)

		cidrSpan = cidrs[1:]
	}

	spew.Dump(subnet)
	fmt.Printf("subenet.Last(): %b\n", subnet.Last().IP)
	fmt.Printf("end: %b\n", end.IP)

	if subnet.Last().GreaterThan(end) {
		scs.Dump(end)
		scs.Dump(subnet.Last())
		excludeAddress := end
		_, err := excludeAddress.Increment(NewIPNumber(1))
		if err != nil {
			return nil, err
		}
		exclude := newNetworkFromIP(version, excludeAddress)
		cidrs = append(cidrs, subnet.Partition(exclude).before...)
	} else {
		cidrs = append(cidrs, cidrSpan...)
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
func (mask *IPMask) Length() *IPNumber {
	ones, bits := net.IPMask(*mask.IPMask).Size()
	return NewIPNumber(2).Exp(NewIPNumber(int64(bits - ones)))
}

func (nw *IPNetwork) Length() *IPNumber { return nw.Mask.Length() }

//// Valid returns true when a subnetwork has a valid mask and start IP.
//func (nw *IPNetwork) Valid() bool {
//	nw.
//	return nw.First().Mask(nw.Mask()) == nw.start
//}
//
//func (num IPNumber) Mask(mask CIDRMask) *IPNumber {
//	return IPAddress(net.IP(ipNum.ToIPAddress()).Mask(net.IPMask(mask))).ToInt()
//}
