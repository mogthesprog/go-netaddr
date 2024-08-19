package netaddr

// IPRange represents a range of IP addresses. It includes the IP version (IPv4 or IPv6),
// the first and last IP addresses in the range, and the network to which the range belongs.
type IPRange struct {
	version *Version
	first   *IPAddress
	last    *IPAddress
	network *IPNetwork
}

// ByIPRanges is a type that implements sort.Interface for sorting a slice of IPRange.
// It sorts the IP ranges first by version (IPv4 or IPv6), then by the starting IP address,
// then by the ending IP address, and finally by the network if the previous criteria are equal.
type ByIPRanges []IPRange

// Len returns the number of IP ranges in the slice. It is required by sort.Interface.
//
// Example usage:
//
//	ranges := netaddr.ByIPRanges{range1, range2, range3}
//	fmt.Println(ranges.Len()) // Output: 3
func (rs ByIPRanges) Len() int {
	return len(rs)
}

// Less reports whether the IP range at index i should sort before the IP range at index j.
// It first compares the IP versions, then the start addresses, then the end addresses,
// and finally the network if the previous comparisons are equal.
//
// Example usage:
//
//	ranges := netaddr.ByIPRanges{range1, range2}
//	sort.Sort(ranges)
//	fmt.Println(ranges)
func (rs ByIPRanges) Less(i, j int) bool {
	ith := rs[i]
	jth := rs[j]
	if ith.version != jth.version {
		return ith.version.LessThan(jth.version)
	}
	if !ith.first.Equal(jth.first) {
		return ith.first.LessThan(jth.first)
	}
	if !ith.last.Equal(jth.last) {
		return ith.last.LessThan(jth.last)
	}
	if !ith.network.Equal(jth.network) {
		return ith.network.LessThan(jth.network)
	}
	return false
}

// Swap exchanges the IP ranges at indices i and j. It is required by sort.Interface.
//
// Example usage:
//
//	ranges := netaddr.ByIPRanges{range1, range2}
//	ranges.Swap(0, 1)
//	fmt.Println(ranges)
func (rs ByIPRanges) Swap(i, j int) {
	ith := rs[i]
	rs[i] = rs[j]
	rs[j] = ith
}
