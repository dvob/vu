package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/bytefmt"
)

// func bindFlags(cmd *cobra.Command) {
// 	uri := ""
// 	cmd.Flags().StringVar(&uri, "uri", "unix:/var/run/libvirt/libvirt-sock", "libvirt listen address. either in unix:/socket/path or tcp:127.0.0.1 format")
// }

func errPrint(e ...interface{}) {
	fmt.Fprintln(os.Stderr, e...)
}

func errExit(e ...interface{}) {
	errPrint(e...)
	os.Exit(1)
}

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
