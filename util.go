package main

import (
	"github.com/cloudfoundry/bytefmt"
)

// ByteSize implements the Value interface for flag parsing with pflag
type ByteSize uint64

func (b *ByteSize) String() string {
	return bytefmt.ByteSize(uint64(*b))
}

// Set sets the value
func (b *ByteSize) Set(value string) (err error) {
	bval, err := bytefmt.ToBytes(value)
	*b = ByteSize(bval)
	return err
}

// Type returns the type of ByteSize
func (b *ByteSize) Type() string {
	return "uint64"
}
