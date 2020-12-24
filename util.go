package main

import (
	"code.cloudfoundry.org/bytefmt"
)

// ByteSize implements the Value interface for flag parsing with pflag
type ByteSize struct {
	val *uint64
}

func NewByteSize(bytes *uint64) *ByteSize {
	return &ByteSize{
		val: bytes,
	}
}

func (b *ByteSize) String() string {
	return bytefmt.ByteSize(uint64(*b.val))
}

// Set sets the value
func (b *ByteSize) Set(value string) (err error) {
	bval, err := bytefmt.ToBytes(value)
	*b.val = bval
	return err
}

// Type returns the type of ByteSize
func (b *ByteSize) Type() string {
	return "uint64"
}
