package netaddr

type IPRange struct {
	version *Version
	first   *IPAddress
	last    *IPAddress
	network *IPNetwork
}

// Implements sort.Interface for sorting IPRange
type ByIPRanges []IPRange

func (rs ByIPRanges) Len() int {
	return len(rs)
}

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

func (rs ByIPRanges) Swap(i, j int) {
	ith := rs[i]
	rs[i] = rs[j]
	rs[j] = ith
}
