package internal

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/digitalocean/go-libvirt"
	image "github.com/dvob/vu/internal/image/libvirt"
	vm "github.com/dvob/vu/internal/vm/libvirt"
	"github.com/spf13/cobra"
)

type LibvirtOptions struct {
	URI          string
	BaseImageDir string
}

func (o *LibvirtOptions) BindFlags(cmd *cobra.Command, prefix string) {
	cmd.Flags().StringVar(&o.URI, prefix+"uri", o.URI, "URI to connecto to libvirtd. either a unix socket in the format unix:/socket/path or an IP in the format tcp:127.0.0.1.")
	cmd.Flags().StringVar(&o.BaseImageDir, prefix+"image-base-dir", o.BaseImageDir, "Base directory to create new storage pools for the images.")
}

func NewLibvirtDefaultOptions() *LibvirtOptions {
	return &LibvirtOptions{
		URI:          "unix:/var/run/libvirt/libvirt-sock",
		BaseImageDir: "/var/lib/libvirt/images/vu",
	}
}

func NewLibvirtManager(o *LibvirtOptions) (*Manager, error) {
	libvirtConn, err := connectLibvirt(o.URI)
	if err != nil {
		return nil, err
	}
	return &Manager{
		ConfigImagePool: "config",
		BaseImagePool:   "base",
		VMImagePool:     "vm",
		Image:           image.New(o.BaseImageDir, libvirtConn),
		VM:              vm.New(libvirtConn),
	}, nil
}

func connectLibvirt(uri string) (*libvirt.Libvirt, error) {
	parts := strings.SplitN(uri, ":", 2)

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid connection uri '%s'", uri)
	}

	network := parts[0]
	address := parts[1]
	d := newDialer(network, address, 2*time.Second)

	libvirtConn := libvirt.NewWithDialer(d)
	if err := libvirtConn.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	return libvirtConn, nil
}

type dialer struct {
	network string
	address string
	timeout time.Duration
}

func newDialer(network, address string, timeout time.Duration) *dialer {
	return &dialer{
		network: network,
		address: address,
		timeout: timeout,
	}
}

func (d *dialer) Dial() (net.Conn, error) {
	return net.DialTimeout(d.network, d.address, d.timeout)
}
