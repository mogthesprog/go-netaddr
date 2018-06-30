package netaddr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPRangeToCIDRs(t *testing.T) {
	t.Parallel()

	networks, err := IPRangeToCIDRS(IPv4, NewIP("0.0.0.0"), NewIP("255.255.255.255"))
	assert.NoError(t, err)

	assert.Equal(t, networks, networks)

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
