package netaddr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//func (ip *IPAddress) ToInt() *IPNumber {
//	num := NewIPNumber(0)
//	num.SetBytes(*ip.IP)
//	return num
//}
//
//func (num *IPNumber) ToIPAddress() *IPAddress {
//	bytes := make(net.IP, 16)
//	bigintBytes := num.Bytes()
//	for i := 0; i < len(bigintBytes); i++ {
//		[]byte(bytes)[len(bytes)-(i+1)] = bigintBytes[len(bigintBytes)-(i+1)]
//	}
//	return &IPAddress{IP: &bytes}
//}

func TestIPAddressToIntConversion(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		addr *IPAddress
	}{
		{NewIP("0.0.0.0")},
		{NewIP("255.255.255.255")},
		{NewIP("1.1.1.1")},
	}

	for _, test := range tests {
		ipFromInt := test.addr.ToInt().ToIPAddress()

		assert.NotNil(t, test.addr.version)
		assert.NotNil(t, ipFromInt.version)
		assert.Equal(t, test.addr, ipFromInt)
	}

}
