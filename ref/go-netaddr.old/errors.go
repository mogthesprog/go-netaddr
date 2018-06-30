package netaddr

import "fmt"

var (
	ErrNoSpaceLeftInNetwork = fmt.Errorf("no space left in go-netaddr for subnet")
)
