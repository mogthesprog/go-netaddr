package netaddr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestIncrement(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		initialValue *IPAddress
		incrementBy  int64
		expected     *IPAddress
		expectedError error
	}{
		{ NewIP("1.1.1.1"), 1, NewIP("1.1.1.2"), nil },
		{ NewIP("1.1.1.1"), 5, NewIP("1.1.1.6"), nil },
		{ NewIP("1.1.1.255"), 1, NewIP("1.1.2.0"), nil },
		{ NewIP("1.1.1.254"), 3, NewIP("1.1.2.1"), nil },
		{ NewIP("255.255.255.255"), 1, nil, ErrorAddressOutOFBounds },
	}

	for _, test := range tests {
		result, err := test.initialValue.Increment(NewIPNumber(test.incrementBy))
		assert.Equal(t, test.expected, result)
		assert.Equal(t, test.expectedError, err)
	}

}
