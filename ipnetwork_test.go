package netaddr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPRangeToCIDRs(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		start *IPAddress
		end   *IPAddress
		exp   []*IPNetwork
	}{
		{
			NewIP("1.1.1.0"),
			NewIP("1.1.1.255"),
			[]*IPNetwork{
				newTestNetwork(t, "1.1.1.0/24"),
			},
		},
		{
			NewIP("1.1.1.0"),
			NewIP("1.1.2.255"),
			[]*IPNetwork{
				newTestNetwork(t, "1.1.1.0/24"),
				newTestNetwork(t, "1.1.2.0/24"),
			},
		},
		{
			NewIP("0.0.0.0"),
			NewIP("10.255.255.25"),
			[]*IPNetwork{
				newTestNetwork(t, "0.0.0.0/5"), newTestNetwork(t, "8.0.0.0/7"),
				newTestNetwork(t, "10.0.0.0/9"), newTestNetwork(t, "10.128.0.0/10"),
				newTestNetwork(t, "10.192.0.0/11"), newTestNetwork(t, "10.224.0.0/12"),
				newTestNetwork(t, "10.240.0.0/13"), newTestNetwork(t, "10.248.0.0/14"),
				newTestNetwork(t, "10.252.0.0/15"), newTestNetwork(t, "10.254.0.0/16"),
				newTestNetwork(t, "10.255.0.0/17"), newTestNetwork(t, "10.255.128.0/18"),
				newTestNetwork(t, "10.255.192.0/19"), newTestNetwork(t, "10.255.224.0/20"),
				newTestNetwork(t, "10.255.240.0/21"), newTestNetwork(t, "10.255.248.0/22"),
				newTestNetwork(t, "10.255.252.0/23"), newTestNetwork(t, "10.255.254.0/24"),
				newTestNetwork(t, "10.255.255.0/28"), newTestNetwork(t, "10.255.255.16/29"),
				newTestNetwork(t, "10.255.255.24/31"),
			},
		},
		{
			NewIP("0.0.0.0"),
			NewIP("255.255.255.255"),
			[]*IPNetwork{
				newTestNetwork(t, "0.0.0.0/0"),
			},
		},
	}

	for _, test := range tests {
		subnets, err := IPRangeToCIDRS(IPv4, test.start, test.end)
		assert.NoError(t, err)

		scs.Dump(subnets)

		assert.Equal(t, test.exp, subnets)

	}

}

func TestIPNetworkPartition(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		target *IPNetwork
		exclude  *IPNetwork
		expected Partition
	} {
		{
			target: newTestNetwork(t, "1.1.2.0/23"),
			exclude: newTestNetwork(t, "1.1.3.0/32"),
			expected: Partition{
				before: []*IPNetwork{
					newTestNetwork(t, "1.1.2.0/24"),
				},
				partition: newTestNetwork(t, "1.1.3.0/32"),
				after: []*IPNetwork{
					newTestNetwork(t, "1.1.3.1/32"), newTestNetwork(t, "1.1.3.2/31"),
					newTestNetwork(t, "1.1.3.4/30"), newTestNetwork(t, "1.1.3.8/29"),
					newTestNetwork(t, "1.1.3.16/28"), newTestNetwork(t, "1.1.3.32/27"),
					newTestNetwork(t, "1.1.3.64/26"), newTestNetwork(t, "1.1.3.128/25"),
				},
			},
		},
		{
			target: newTestNetwork(t, "1.1.0.0/22"),
			exclude: newTestNetwork(t, "1.1.0.255/32"),
			expected: Partition{
				before: []*IPNetwork{
					newTestNetwork(t, "1.1.0.0/25"), newTestNetwork(t, "1.1.0.128/26"),
					newTestNetwork(t, "1.1.0.192/27"), newTestNetwork(t, "1.1.0.224/28"),
					newTestNetwork(t, "1.1.0.240/29"), newTestNetwork(t, "1.1.0.248/30"),
					newTestNetwork(t, "1.1.0.252/31"), newTestNetwork(t, "1.1.0.254/32"),
				},
				partition: newTestNetwork(t, "1.1.0.255/32"),
				after: []*IPNetwork{newTestNetwork(t, "1.1.1.0/24"),newTestNetwork(t, "1.1.2.0/23")},
			},
		},
	}

	for _, test := range tests {
		result := *test.target.Partition(test.exclude)
		assert.Equal(t, test.expected, result)
	}
}

func TestNewNetworkFromIP(t *testing.T) {
	nw := newNetworkFromIP(IPv4, NewIP("1.1.1.1"))
	assert.Equal(t, newTestNetwork(t, "1.1.1.1/32"), nw)
}

func newTestNetwork(t *testing.T, net string) *IPNetwork {
	nw, err := NewIPNetwork(net)
	assert.NoError(t, err)
	return nw
}

func newTestNetworkFromBoundaries(t *testing.T, first, last *IPAddress) *IPNetwork {
	nw, err := newNetworkFromBoundaries(first, last)
	assert.NoError(t, err)
	return nw
}

func TestFirstNetworkAddress(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		net *IPNetwork
		exp *IPAddress
	}{
		{newTestNetwork(t, "10.0.0.0/8"), NewIP("10.0.0.0")},
		{newTestNetwork(t, "0.0.0.0/0"), NewIP("0.0.0.0")},
	}

	for _, test := range tests {
		assert.Equal(t, test.exp, test.net.First())
	}
}

func TestLastNetworkAddress(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		net *IPNetwork
		exp *IPAddress
	}{
		{newTestNetwork(t, "10.0.0.0/8"), NewIP("10.255.255.255")},
		{newTestNetwork(t, "0.0.0.0/0"), NewIP("255.255.255.255")},
	}

	for _, test := range tests {
		assert.Equal(t, test.exp, test.net.Last())
	}
}

func TestNewIPNetwork(t *testing.T) {
	t.Parallel()
	nw, err := NewIPNetwork("10.0.0.0/8")
	assert.NoError(t, err)
	assert.Equal(t, IPv4, nw.version)
	assert.Equal(t, NewIPNumber(8), nw.PrefixLength())
	assert.Equal(t, NewMask(8, 32), nw.Mask)
}

func TestNetworkLength(t *testing.T) {
	t.Parallel()
	nw, err := NewIPNetwork("10.0.0.0/8")
	assert.NoError(t, err)
	assert.Equal(t, NewIPNumber(16777216), nw.Length())
}

func TestIPNetworkEqual(t *testing.T) {
	t.Parallel()

	network1, _ := NewIPNetwork("10.0.0.0/8")
	network2, _ := NewIPNetwork("10.0.0.0/8")
	network3, _ := NewIPNetwork("10.0.0.1/8")
	network4, _ := NewIPNetwork("10.0.0.1/16")
	var tests = []struct {
		name          string
		firstNetwork  *IPNetwork
		secondNetwork *IPNetwork
		expected      bool
	}{
		{"Same object is equal", network1, network1, true},
		{"Different object same network is equal", network1, network2, true},
		{"Different IP same mask is same", network1, network3, true},
		{"Same IP different mask is different", network3, network4, false},
	}

	for _, test := range tests {
		result := test.firstNetwork.Equal(test.secondNetwork)
		assert.Equal(t, test.expected, result, "%v: IPNetwork.Equal() = %v, want %v", test.name, result, test.expected)
	}
}

func TestIPNetworkLessThan(t *testing.T) {
	t.Parallel()

	network1, _ := NewIPNetwork("10.0.0.0/24")
	network2, _ := NewIPNetwork("10.0.0.0/24")
	network3, _ := NewIPNetwork("10.0.1.0/24")
	network4, _ := NewIPNetwork("10.0.0.0/25")
	network5, _ := NewIPNetwork("10.0.0.0/24")
	var tests = []struct {
		name          string
		firstNetwork  *IPNetwork
		secondNetwork *IPNetwork
		expected      bool
	}{
		{"Same network is not less than", network1, network2, false},
		{"Lower IP same mask", network1, network3, true},
		{"Higher IP same mask", network3, network1, false},
		{"Same IP lower mask", network5, network4, true},
		{"Same IP higher mask", network4, network5, false},
	}

	for _, test := range tests {
		result := test.firstNetwork.LessThan(test.secondNetwork)
		assert.Equal(t, test.expected, result, "%v: IPNetwork.LessThan() = %v, want %v", test.name, result, test.expected)
	}
}

func TestNewMask(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		net *IPNetwork
		exp *IPMask
	}{
		{
			newTestNetwork(t, "10.0.0.0/8"),
			NewMask(8, 32),
		},
		{
			newTestNetwork(t, "0.0.0.0/0"),
			NewMask(0, 32),
		},
	}
	for _, test := range tests {

		assert.Equal(t, test.exp, test.net.Mask)
	}

}

func TestNewNetworkFromBoundaries(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		net *IPNetwork
		exp *IPNetwork
	}{
		{
			newTestNetworkFromBoundaries(t, NewIP("10.0.0.0"), NewIP("10.255.255.255")),
			newTestNetwork(t, "10.0.0.0/8"),
		},
		{
			newTestNetworkFromBoundaries(t, NewIP("0.0.0.0"), NewIP("0.255.255.255")),
			newTestNetwork(t, "0.0.0.0/8"),
		},
		{
			newTestNetworkFromBoundaries(t, NewIP("0.0.0.0"), NewIP("255.255.255.255")),
			newTestNetwork(t, "0.0.0.0/0"),
		},
	}

	for _, test := range tests {

		assert.Equal(t, test.exp, test.net, "error creating network: %s", test.exp)
	}
}
