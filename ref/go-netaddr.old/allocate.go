package netaddr

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"net"
)

var (
	DefaultNetwork = []net.IPNet{
		{net.ParseIP("10.0.0.0"), net.CIDRMask(8, 32)},
		{net.ParseIP("172.16.0.0"), net.CIDRMask(12, 32)},
		{net.ParseIP("192.168.0.0"), net.CIDRMask(16, 32)},
	}
)

type (
	// IPNumber is the uint32 representation of an IPv4
	IPNumber uint32
	// IPNetwork is an abstraction over the net.IPNet type in the standard library
	IPNetwork net.IPNet
	// IPaddress is an abstraction over the net.IP type in the standard library
	IPAddress net.IP
	// CIDRMask is the Subnet mask in CIDR notation
	//
	// i.e. for a subnet 127.0.0.1/8 - CIDRMask = 8
	CIDRMask net.IPMask
)

// Network is blah
//
// Network is not thread safe.
type Network struct {
	allocatable []*Subnetwork
	allocated   []*Subnetwork
}

type Subnetwork struct {
	start IPNumber
	mask  CIDRMask
}

// internal type for planning allocation of a subnetwork
type allocationPlan struct {
	allocated *Subnetwork
	remaining []*Subnetwork
}

// NewNetwork returns a Network populated with all available cidrs... as allocatable.
// Optionally you can pass in DefaultNetwork in place of cidrs... to have all RFC1918
// address space made available.
func NewNetwork(cidrs ...net.IPNet) *Network {
	var allocatable = []*Subnetwork{}
	for _, v := range cidrs {
		allocatable = append(allocatable, CIDRToSubnetwork(v))
	}
	return &Network{
		allocatable: allocatable,
		allocated:   []*Subnetwork{},
	}
}

func NewSubnetwork(first IPAddress, mask CIDRMask) (*Subnetwork, error) {
	subnetwork := &Subnetwork{
		start: first.ToInt(),
		mask:  mask,
	}
	if !subnetwork.Valid() {
		return nil, fmt.Errorf("subnet %v is not a valid CIDR", subnetwork)
	}
	return subnetwork, nil
}

func (s *Subnetwork) Length() int { return s.mask.Length() }

// Valid returns true when a subnetwork has a valid mask and start IP.
func (s *Subnetwork) Valid() bool {
	return s.start.Mask(s.mask) == s.start
}

func (ipNum IPNumber) Mask(mask CIDRMask) IPNumber {
	return IPAddress(net.IP(ipNum.ToIPAddress()).Mask(net.IPMask(mask))).ToInt()
}

// CIDRToSubnetwork converts a given net.IPNet to a Subnetwork.
func CIDRToSubnetwork(cidr net.IPNet) *Subnetwork {
	mask := CIDRMask(cidr.Mask)
	return &Subnetwork{
		start: IPToInt(IPAddress(cidr.IP)),
		mask:  mask,
	}
}

// Allocate finds the next available subnetwork of specified mask and updates
// the go-netaddr representation in the project metadata
// Once allocated, the allocated subnetwork is returned, with nil error.
func (n *Network) Allocate(mask CIDRMask) (*Subnetwork, error) {

	// Find the next available subnetwork
	subnetwork, err := n.NextAvailableSubnetwork(mask)
	if err != nil {
		return nil, err
	}

	// Now we can mark the subnet as allocated, and remove it from allocatable
	n.allocate(subnetwork)

	return subnetwork, nil
}

// allocate is an internal method for the allocated and allocatable addresses within Network, n.
func (n *Network) allocate(subnetwork *Subnetwork) error {
	for i, v := range n.allocatable {
		if v.ContainsSubnetwork(subnetwork) {
			plan, err := v.planAllocation(subnetwork)
			if err != nil {
				return err
			}

			// Here we replace the instance of Network.allocatable[i] with the remaining ([]Subnetwork)
			// returned from v.Allocate()
			n.allocatable = append(n.allocatable[:i], append(plan.remaining, n.allocatable[i+1:]...)...)

			// Now we mark the subnetwork as allocated
			n.allocated = append(n.allocated, plan.allocated)

			return nil
		}
	}

	// This code should never get reached
	return fmt.Errorf("an unkown error occured, unable to allocate go-netaddr")
}

// NextAvailableSubnetwork finds the next available Subnetwork of size mask that is also a valid
// CIDR within the Network, n.
func (n *Network) NextAvailableSubnetwork(mask CIDRMask) (*Subnetwork, error) {
	for _, v := range n.allocatable {
		if !(v.Length() > mask.Length()) {
			continue
		}

		maskedSubnetworkNumber := v.start.Mask(mask)
		newSubnetwork, err := NewSubnetwork(maskedSubnetworkNumber.ToIPAddress(), mask)
		if err != nil {
			return nil, err
		}

		if v.ContainsSubnetwork(newSubnetwork) {
			return newSubnetwork, nil
		}

		// in the case v didn't contain newSubnetwork, let's increment and try again
		nextSubnetwork := newSubnetwork.NextSubnetwork()

		if !v.ContainsSubnetwork(newSubnetwork) &&
			v.ContainsSubnetwork(nextSubnetwork) {
			return nextSubnetwork, nil
		}

	}

	// If we've got here then we didn't find an available go-netaddr
	return nil, ErrNoSpaceLeftInNetwork
}

func (s *Subnetwork) NextSubnetwork() *Subnetwork {
	subnetwork := *s
	subnetwork.start = s.start + IPNumber(s.mask.Length())
	return &subnetwork
}

// planAllocation is an internal method used to plan the carving of a subnetwork object into a desired
// subnetwork and any remaining subnetworks which are still allocatable.
//
// planAllocation handles the situation where a request is made for invalid subnets
// i.e. 10.0.1.0/16 -> out of range for a 10.0.0.0/8 go-netaddr
func (s *Subnetwork) planAllocation(subnetwork *Subnetwork) (plan *allocationPlan, err error) {
	plan = &allocationPlan{
		allocated: subnetwork,
		remaining: s.Subtract(subnetwork),
	}

	return
}

// Subtract subtracts a subnetwork from a larger parent subnetwork, returning the slice of remaining networks
// existing outside of the desired subnetwork. Effectively allocating the subnetwork.
func (s *Subnetwork) Subtract(subnetwork *Subnetwork) []*Subnetwork {
	var (
		before *Subnetwork
	)
	// the process might be
	// split the original subnet in two, the preceding and proceeding
	// 1. check the new subnet is valid,
	// 2. if it isn't valid, reduce the subnetCIDR by one and mask again,
	// 3. repeat until valid
	// 4. when the subnet is valid, subtract it from the beginning of the preceding subnet, create a new subnet
	//    with a new representative mask (calculated from the remaining length) and repeat.
	// 5. repeat the above for the proceeding networks too.
	if s.start-subnetwork.start == 0 {
		before = &Subnetwork{}
	} else {
		before := &Subnetwork{s.start, subnetwork.start - s.start}
	}
	rawRemaining := []*Subnetwork{
		{s.start},
	}

	validRemaining := new([]*Subnetwork)

	whilr

	// here we need to calculate the remaining subnetworks. All of them need to be valid, so we may end up with
	// multiple subnets either side of the allocated subnet.
	remaining := []*Subnetwork{
		NewSubnetwork(s.start),
		NewSubnetwork(),
	}

}

func (s *Subnetwork) FirstAddress() IPAddress {
	return s.start.ToIPAddress()
}

func (s *Subnetwork) LastAddress() IPAddress {
	return (s.start + IPNumber(s.Length())).ToIPAddress()

}

func (s *Subnetwork) ContainsAddress(addr IPAddress) bool {
	ipAddressNumber := addr.ToInt()
	return s.start < ipAddressNumber && ipAddressNumber < s.start+IPNumber(s.Length())
}

func (s *Subnetwork) ContainsSubnetwork(subnetwork *Subnetwork) bool {
	return s.start < subnetwork.start &&
		subnetwork.start+IPNumber(subnetwork.Length()) < s.start+IPNumber(s.Length())
}

func (addr IPAddress) ToInt() IPNumber {
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(net.IP(addr).To4())
	return IPNumber(IPv4Int.Uint64())
}

func (number IPNumber) ToIPAddress() IPAddress {
	var ipByte [4]byte
	ip := ipByte[:]
	binary.BigEndian.PutUint32(ip, uint32(number))
	return IPAddress(ip)
}

//
func (mask *CIDRMask) Length() int {
	ones, bits := net.IPMask(*mask).Size()
	return int(math.Pow(2, float64(bits-ones)))
}
