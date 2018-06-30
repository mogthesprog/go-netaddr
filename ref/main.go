package main

import (
	"fmt"
	"math/big"
	"net"
)

type IPAddress net.IP
type IPNumber big.Int

//func (ip IPAddress) ToInt() big.Int {
//	if len(ip) == 16 {
//		return binary.BigEndian.Uint32(ip[12:16])
//	}
//	return binary.BigEndian.Uint32(ip)
//}

//func (number IPNumber) ToIPAddress() IPAddress {
//	ip := make(net.IP, 4)
//	binary.BigEndian.PutUint32(ip, uint32(number))
//	return IPAddress(ip)
//}

func (ip IPAddress) ToInt() IPNumber {
	num := big.NewInt(0)
	num.SetBytes(ip)
	output := IPNumber(*num)
	return output
}

func (number IPNumber) ToIPAddress() IPAddress {
	bytes := make([]byte, 16)
	bigInt := big.Int(number)
	bigintBytes := (bigInt).Bytes()
	for i := 0; i < len(bigintBytes); i++ {
		bytes[len(bytes)-(i+1)] = bigintBytes[len(bigintBytes)-(i+1)]
	}
	return bytes
}

func testEq(a, b []byte) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func main() {
	//340282366920938463463374607431768211455

	beans := IPAddress(net.ParseIP("255.255.255.254"))

	before := []byte(beans)
	after := []byte(beans.ToInt().ToIPAddress())

	fmt.Printf("IP Address: %s\n", net.IP(before))
	fmt.Printf("IPAddress Bytes: %s\n", net.IP(after))

	//num := big.NewInt(0)
	//num.SetBytes(beans)

	//bytes := make([]byte, 16)
	//bigintBytes := num.Bytes()
	//for i, val := range bigintBytes {
	//	//fmt.Println(i, bytes, len(bytes))
	//	bytes[len(bytes)-(i+1)] = val
	//}

	//before = []byte(beans)
	//after = []byte(bytes)

	////fmt.Printf("\n\nIP Address: %s\n", net.IP(beans).To16())
	////fmt.Printf("IP BigInt: %d\n", num)
	////fmt.Printf("IPAddress Bytes: %b\n", net.IP(bytes))

	//	if !testEq(before, after) {
	//		//fmt.Printf("expected %b to equal %b, but failed\n", before, after)
	//	}

	//if 9223372036854775807 < beans {
	//	fmt.Println("Hello, playground")
	//}
}
