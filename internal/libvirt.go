package internal

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/digitalocean/go-libvirt"
	image "github.com/dvob/vu/internal/image/libvirt"
	vm "github.com/dvob/vu/internal/vm/libvirt"
)

type LibvirtOptions struct {
	URI          string
	BaseImageDir string
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
	c, err := net.DialTimeout(network, address, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirtd: %s", err)
	}

	libvirtConn := libvirt.New(c)
	if err := libvirtConn.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	return libvirtConn, nil
}
