// +build integration

package libvirt

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/matryer/is"
)

func conn() (*libvirt.Libvirt, error) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirtd: %s", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	return l, nil
}

func Test_CreateVolume(t *testing.T) {
	is := is.New(t)
	t.Log("foo")
	conn, err := conn()
	if err != nil {
		t.Fatal(err)
	}

	dom, err := conn.DomainLookupByName("net2-dhcp")
	is.NoErr(err)

	for _, i := range []uint32{0, 1, 2, 3} {
		t.Log("source", i)
		ifs, err := conn.DomainInterfaceAddresses(dom, i, 0)
		if err != nil {
			t.Log(err)
		} else {
			printIFs(t, ifs)
		}
	}
}

func printIFs(t *testing.T, ifs []libvirt.DomainInterface) {
	for _, i := range ifs {
		t.Log(i)
	}
}
