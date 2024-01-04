package main

import (
	"fmt"
)

type IP struct {
	// Subnet
	Subnet int
	// Host
	Host int
}

func (ip IP) String() string {
	return fmt.Sprintf("192.168.%d.%d", ip.Subnet, ip.Host)
}

func (ip IP) MarshalText() ([]byte, error) {
	return []byte(ip.String()), nil
}

func (ip *IP) UnmarshalText(b []byte) error {
	fmt.Sscanf(string(b), "192.168.%d.%d", &ip.Subnet, &ip.Host)
	return nil
}
