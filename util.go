package main

import (
	"github.com/cloudfoundry/bytefmt"
)

// byte size implements value interface
type ByteSize uint64

func (b ByteSize) String() string {
	return bytefmt.ByteSize(uint64(b))
}

func (b ByteSize) Set(value string) (err error) {
	btmp, err := bytefmt.ToBytes(value)
	b = ByteSize(btmp)
	return err
}

//TODO: what should Type implement?
func (b ByteSize) Type() string {
	return "uint64"
}
