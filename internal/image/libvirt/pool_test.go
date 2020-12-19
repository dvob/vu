package libvirt

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/digitalocean/go-libvirt"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
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
	t.Log("foo")
	conn, err := conn()
	if err != nil {
		t.Fatal(err)
	}

	sp, err := conn.StoragePoolLookupByName("cloud_image")
	if err != nil {
		t.Fatal(err)
	}

	vol := &libvirtxml.StorageVolume{
		Name: "mytest",
		Capacity: &libvirtxml.StorageVolumeSize{
			Value: 0,
		},
		// Target: &libvirtxml.StorageVolumeTarget{
		// 	Format: &libvirtxml.StorageVolumeTargetFormat{
		// 		Type: kind,
		// 	},
		// },
	}

	xml, err := vol.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	sv, err := conn.StorageVolCreateXML(sp, xml, 0)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	buf.WriteString("hello world!")

	err = conn.StorageVolUpload(sv, &buf, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

}
