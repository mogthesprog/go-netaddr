# netaddr

A library for network address manipulation, written in Go.

This library is a Go port of the Python netaddr package. It provides a means of converting IP Addresses to integers and back again, including tools for manipulating CIDRs and performing basic comparisons of CIDRs and Addresses.

[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/mogthesprog/go-netaddr.svg)](https://github.com/mogthesprog/go-netaddr)
[![Go Report Card](https://goreportcard.com/badge/github.com/mogthesprog/go-netaddr)](https://goreportcard.com/report/github.com/mogthesprog/go-netaddr)
[![Go Reference](https://pkg.go.dev/badge/github.com/mogthesprog/go-netaddr.svg)](https://pkg.go.dev/github.com/mogthesprog/go-netaddr)
[![Test Status](https://github.com/mogthesprog/go-netaddr/actions/workflows/actions.yaml/badge.svg)](https://github.com/mogthesprog/netaddr/actions/workflows/actions.yaml)

## Table of Contents
- [Installation](#installation)
- [Usage](#usage)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## Installation

To install the `go-netaddr` library, use the following command:

```bash
go get github.com/mogthesprog/go-netaddr
```

## Usage

Here's a simple example of how to use the `go-netaddr` library:

```go
package main

import (
    "fmt"
    "github.com/mogthesprog/go-netaddr"
)

func main() {
    ip, _ := netaddr.ParseIP("192.168.0.1")
    fmt.Println(ip.ToInt())
    
    cidr, _ := netaddr.ParseCIDR("192.168.0.0/24")
    fmt.Println(cidr.Contains(ip))
}
```

This example demonstrates basic IP address parsing, conversion to an integer, and CIDR manipulation. The library also supports more advanced features such as IP range calculations and subnet comparisons.

## Documentation

Full documentation is available on [GoDoc](https://pkg.go.dev/github.com/mogthesprog/go-netaddr).

## Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) file for details on how to contribute to this project.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
