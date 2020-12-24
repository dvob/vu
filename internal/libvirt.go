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
	URI               string
	ConfigStorageName string
	ConfigStoragePath string
	ImageStorageName  string
	ImageStoragePath  string
}

func NewLibvirtDefaultOptions() *LibvirtOptions {
	return &LibvirtOptions{
		URI:               "unix:/var/run/libvirt/libvirt-sock",
		ImageStorageName:  "vu_images",
		ImageStoragePath:  "/var/lib/libvirt/images/vu/images",
		ConfigStorageName: "vu_configs",
		ConfigStoragePath: "/var/lib/libvirt/images/vu/configs",
	}

}

func NewLibvirtManager(o *LibvirtOptions) (*Manager, error) {
	libvirtConn, err := connectLibvirt(o.URI)
	if err != nil {
		return nil, err
	}
	return &Manager{
		ConfigImage: image.New(o.ConfigStorageName, o.ConfigStoragePath, libvirtConn),
		BaseImage:   image.New(o.ImageStorageName, o.ImageStoragePath, libvirtConn),
		VM:          vm.New(libvirtConn),
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
