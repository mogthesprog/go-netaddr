// Package netaddr provides utilities for manipulating IP addresses and CIDRs.
// It offers functionality to manage subnet ranges, determine whether IPs are
// included in CIDRs, split CIDRs into subranges, and perform various other
// operations related to IP addressing and network calculations.
//
// This library is a go port of the Python netaddr package, a lot of the credit
// for the existence of this package goes to the maintainers and contributors of
// python netaddr.
//
// # Overview
//
// This package is designed for network engineers, developers, and system
// administrators who need to handle IP address operations programmatically.
// It supports both IPv4 and IPv6 addresses and provides a range of tools to
// simplify working with IP addresses, CIDR blocks, and IP ranges.
//
// # Features
//
// - Parse and manipulate individual IP addresses.
// - Create and manipulate CIDR blocks (IP networks).
// - Determine if an IP address belongs to a specific CIDR block.
// - Split a CIDR block into smaller subnets.
// - Calculate the range of IP addresses within a CIDR block.
// - Perform operations on IP ranges.
//
// # Getting Started
//
// To start using the package, import it into your Go application:
//
//	import "github.com/mogthesprog/netaddr"
//
// Below are some examples to help you get started:
//
// ## Parsing an IP Address
//
// You can parse an IP address using the NewIPAddress function:
//
//	ip, err := netaddr.NewIPAddress("192.168.1.1")
//	if err != nil {
//		// handle error
//	}
//
// ## Working with CIDR Blocks
//
// You can create a new CIDR block using the NewIPNetwork function:
//
//	network, err := netaddr.NewIPNetwork("192.168.1.0/24")
//	if err != nil {
//		// handle error
//	}
//
// To check if an IP address belongs to this CIDR block:
//
//	isInRange := network.Contains(ip)
//
// ## Splitting a CIDR Block into Subnets
//
// To split a CIDR block into smaller subnets:
//
//	subnets, err := network.SplitIntoSubnets(4)
//	if err != nil {
//		// handle error
//	}
//	for _, subnet := range subnets {
//		fmt.Println(subnet)
//	}
//
// ## Working with IP Ranges
//
// You can create an IP range and perform operations on it:
//
//	ipRange := netaddr.NewIPRange(netaddr.NewIPAddress("192.168.1.1"), netaddr.NewIPAddress("192.168.1.254"))
//
// To check if an IP address is within this range:
//
//	isInRange := ipRange.Contains(ip)
//
// # License
//
// This package is licensed under the Apache 2.0 License. See the LICENSE file for details.
//
// # Contributing
//
// Contributions are welcome! Please see the README file for guidelines on how to contribute.
//
// # Acknowledgments
//
// This package was inspired by the need for a robust and flexible IP manipulation library
// in Go.
package netaddr
