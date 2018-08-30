package netaddr

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	cidrIpv4_1, _ = NewIPNetwork("10.0.0.0/24")
	cidrIpv4_2, _ = NewIPNetwork("10.0.0.0/25")
	cidrIpv6_1, _ = NewIPNetwork("2001:db8::ff00:42:8328/31")
	cidrIpv6_2, _ = NewIPNetwork("2001:db8::ff00:42:8328/32")
	ipv4Range1    = IPRange{IPv4, NewIP("10.0.0.1"), NewIP("10.0.0.2"), cidrIpv4_1}
	ipv4Range2    = IPRange{IPv4, NewIP("10.0.0.3"), NewIP("10.0.0.4"), cidrIpv4_1}
	ipv4Range3    = IPRange{IPv4, NewIP("10.0.0.1"), NewIP("10.0.0.3"), cidrIpv4_2}
	ipv4Range4    = IPRange{IPv4, NewIP("10.0.0.1"), NewIP("10.0.0.3"), cidrIpv4_1}
	ipv6Range1    = IPRange{IPv6, NewIP("2001:db8::ff00:42:8328"), NewIP("2001:db8::ff00:42:8328"), cidrIpv6_1}
	ipv6Range2    = IPRange{IPv6, NewIP("2001:db8::ff00:42:8329"), NewIP("2001:db8::ff00:42:8329"), cidrIpv6_2}
	ipv6Range3    = IPRange{IPv6, NewIP("2001:db8::ff00:42:8328"), NewIP("2001:db8::ff00:42:8329"), cidrIpv6_2}
	ipv6Range4    = IPRange{IPv6, NewIP("2001:db8::ff00:42:8328"), NewIP("2001:db8::ff00:42:8329"), cidrIpv6_1}
)

func TestMain(m *testing.M) {
	returnCode := m.Run()
	os.Exit(returnCode)
}

func TestByIPRanges_Len(t *testing.T) {


	tests := []struct {
		name   string
		ranges ByIPRanges
		want   int
	}{
		{"Empty collection", []IPRange{}, 0},
		{"One element", []IPRange{ipv4Range1}, 1},
		{"Multiple elements", []IPRange{ipv4Range1, ipv4Range2,}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ranges.Len(); got != tt.want {
				t.Errorf("ByIPRanges.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestByIPRanges_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name   string
		ranges ByIPRanges
		args   args
		want   bool
	}{
		{"IPv4 is less than IPv6", ByIPRanges{ipv4Range1, ipv6Range1}, args{0, 1}, true},
		{"IPv6 is not less than IPv4", ByIPRanges{ipv6Range1, ipv4Range1}, args{0, 1}, false},
		{"IPv6 addresses with different first addresses give less than as true correctly based on first", ByIPRanges{ipv6Range1, ipv6Range2}, args{0, 1}, true},
		{"IPv6 addresses with different first addresses give less than as false correctly based on first", ByIPRanges{ipv6Range2, ipv6Range1}, args{0, 1}, false},
		{"IPv4 addresses with different first addresses give less than as true correctly based on first", ByIPRanges{ipv4Range1, ipv4Range2}, args{0, 1}, true},
		{"IPv4 addresses with different first addresses give less than as false correctly based on first", ByIPRanges{ipv4Range2, ipv4Range1}, args{0, 1}, false},
		{"IPv6 addresses with same first addresses give less than as true correctly based on last", ByIPRanges{ipv6Range1, ipv6Range3}, args{0, 1}, true},
		{"IPv6 addresses with same first addresses give less than as false correctly based on last", ByIPRanges{ipv6Range3, ipv6Range1}, args{0, 1}, false},
		{"IPv4 addresses with same first addresses give less than as true correctly based on last", ByIPRanges{ipv4Range1, ipv4Range3}, args{0, 1}, true},
		{"IPv4 addresses with same first addresses give less than as false correctly based on last", ByIPRanges{ipv4Range3, ipv4Range1}, args{0, 1}, false},

		{"IPv6 addresses with same first and last addresses give less than as true correctly based on CIDR", ByIPRanges{ipv6Range4, ipv6Range3}, args{0, 1}, true},
		{"IPv6 addresses with same first and last addresses give less than as false correctly based on CIDR", ByIPRanges{ipv6Range3, ipv6Range4}, args{0, 1}, false},
		{"IPv4 addresses with same first and last addresses give less than as true correctly based on CIDR", ByIPRanges{ipv4Range4, ipv4Range3}, args{0, 1}, true},
		{"IPv4 addresses with same first and last addresses give less than as false correctly based on CIDR", ByIPRanges{ipv4Range3, ipv4Range4}, args{0, 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ranges.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("ByIPRanges.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestByIPRanges_Swap(t *testing.T) {
	ranges := ByIPRanges{ipv4Range1, ipv4Range2}
	expectedRanges := ByIPRanges{ipv4Range2, ipv4Range1}
	ranges.Swap(0, 1)
	assert.Equal(t, expectedRanges, ranges)
}
