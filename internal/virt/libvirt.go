package virt

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"net/http"
	"io"
	"os"
	"time"

	"gopkg.in/cheggaaa/pb.v1"
	"github.com/digitalocean/go-libvirt"
	"github.com/libvirt/libvirt-go-xml"
	"github.com/dsbrng25b/cis/internal/cloud-init"
)

type Manager interface {
	CreateBaseImage(name string, src string) error
	Create(name string, baseImage string) error
	Remove(name string) error
}

type LibvirtManager struct {
	l *libvirt.Libvirt
	pool string
	network string
}

func NewLibvirtManager(pool string, network string) (*LibvirtManager, error) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt sock: %s", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return &LibvirtManager{
		l,
		pool,
		network,
	}, nil
}

func (m *LibvirtManager) removeVolume(name string) error {
	sp, err := m.l.StoragePoolLookupByName(m.pool)
	if err != nil {
		return fmt.Errorf("faild to get storage pool: %s", err)
	}
	sv, err := m.l.StorageVolLookupByName(sp, name)
	if err != nil {
		return err
	}
	err = m.l.StorageVolDelete(sv, 0)
	return err
}


func (m *LibvirtManager) removeDomain(name string) error {
	dom, err := m.l.DomainLookupByName(name)
	if err != nil {
		return err
	}

	stateInt, _, err := m.l.DomainGetState(dom, 0)
	state := libvirt.DomainState(stateInt)
	if err != nil {
		return err
	}
	if state == libvirt.DomainRunning || state == libvirt.DomainPaused {
		err = m.l.DomainDestroy(dom)
		if err != nil {
			return err
		}
	}

	err = m.l.DomainUndefine(dom)
	if err != nil {
		return err
	}

	return nil
}

// creates a volume and uploads the image from the url src into the volume
func (m *LibvirtManager) CreateBaseImage(name string, src string) error {
	var size uint64
	var stream io.Reader
	sp, err := m.l.StoragePoolLookupByName(m.pool)
	if err != nil {
		return fmt.Errorf("faild to get storage pool: %s", err)
	}

	u, err := url.Parse(src)

	if err != nil {
		return fmt.Errorf("failed to parse url: %s", err)
	}

	if u.Scheme == "file" {
		file, err := os.Open(u.Path)

		if err != nil {
			return err
		}

		fileinfo, err := file.Stat()
		if err != nil {
			return err
		}

		if fileinfo.Size() < 0 {
			return fmt.Errorf("negative file size")
		}

		size = uint64(fileinfo.Size())
		stream = file

	} else if u.Scheme == "http" || u.Scheme == "https" {
		resp, err := http.Get(u.String())

		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("http status %d returned", resp.StatusCode)
		}

		if resp.ContentLength < 0 {
			return fmt.Errorf("could not determine content length")
		}

		size = uint64(resp.ContentLength)
		stream = resp.Body

	} else {
		return fmt.Errorf("unkown schema '%s'", u.Scheme)
	}

	vol := &libvirtxml.StorageVolume{
		Name: name,
		Capacity: &libvirtxml.StorageVolumeSize{
			Value: size,
		},
	}

	xml, err := vol.Marshal()
	if err != nil {
		return err
	}

	sv, err := m.l.StorageVolCreateXML(sp, xml, 0)
	if err != nil {
		return fmt.Errorf("failed to create volume: %s", err)
	}

	bar := pb.New(int(size)).SetUnits(pb.U_BYTES)
	bar.Start()

	stream = bar.NewProxyReader(stream)

	err = m.l.StorageVolUpload(sv, stream, 0, 0, 0)
	bar.Finish()
	if err != nil {
		m.l.StorageVolDelete(sv, 0)
		return fmt.Errorf("failed to upload content: %s", err)
	}
	return nil
}

func (m *LibvirtManager) createIsoVolume(name string, size uint64, stream io.Reader) (*libvirt.StorageVol, error) {
	vol := &libvirtxml.StorageVolume{
		Name: name,
		Capacity: &libvirtxml.StorageVolumeSize{
			Value: size,
		},
		Target: &libvirtxml.StorageVolumeTarget{
			Format: &libvirtxml.StorageVolumeTargetFormat{
				Type: "iso",
			},
		},
	}

	xml, err := vol.Marshal()
	if err != nil {
		return nil, err
	}

	sp, err := m.l.StoragePoolLookupByName(m.pool)
	if err != nil {
		return nil, fmt.Errorf("faild to get storage pool: %s", err)
	}

	sv, err := m.l.StorageVolCreateXML(sp, xml, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %s", err)
	}

	err = m.l.StorageVolUpload(sv, stream, 0, 0, 0)
	if err != nil {
		m.l.StorageVolDelete(sv, 0)
		return nil, fmt.Errorf("failed to upload content: %s", err)
	}
	return &sv, nil
}

func (m *LibvirtManager) createVm(name string, baseImage string, configVol libvirt.StorageVol) error {
	sv, err := m.cloneBaseImage(name, baseImage)
	if err != nil {
		return fmt.Errorf("failed to clone %s: %s", baseImage, err)
	}

	path, err := m.l.StorageVolGetPath(*sv)
	if err != nil {
		return err
	}
	configPath, err := m.l.StorageVolGetPath(configVol)
	if err != nil {
		return err
	}
	domain := &libvirtxml.Domain{
		Name: name,
		Type: "kvm",
		Memory: &libvirtxml.DomainMemory{
			Value: 1024,
			Unit: "M",
		},
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Type: "hvm",
			},
		},
		Devices: &libvirtxml.DomainDeviceList{
			Disks: []libvirtxml.DomainDisk{
				libvirtxml.DomainDisk{
					Device: "disk",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "qcow2",
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "hda",
						Bus: "ide",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: path,
						},
					},
				},{
					Device: "cdrom",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "raw",
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "hdb",
						Bus: "ide",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: configPath,
						},
					},
				},
			},
			Interfaces: []libvirtxml.DomainInterface{
				libvirtxml.DomainInterface{
					Source: &libvirtxml.DomainInterfaceSource{
						Network: &libvirtxml.DomainInterfaceSourceNetwork{
							Network: m.network,
						},
					},
				},
			},
			Serials: []libvirtxml.DomainSerial{
				libvirtxml.DomainSerial{},
			},
		},
	}

	xml, err := domain.Marshal()
	if err != nil {
		return nil
	}

	dom, err := m.l.DomainDefineXML(xml)
	if err != nil {
		return err
	}

	err = m.l.DomainCreate(dom)

	return err
}

func (m *LibvirtManager) cloneBaseImage(name string, baseImage string) (*libvirt.StorageVol, error) {
	sp, err := m.l.StoragePoolLookupByName(m.pool)
	if err != nil {
		return nil, fmt.Errorf("faild to get storage pool: %s", err)
	}
	baseImgVol, err := m.l.StorageVolLookupByName(sp, baseImage)
	if err != nil {
		return nil, err
	}
	baseImgPath, err := m.l.StorageVolGetPath(baseImgVol)
	if err != nil {
		return nil, err
	}
	vol := &libvirtxml.StorageVolume{
		Name: name,
		// Capacity: &libvirtxml.StorageVolumeSize{
		// 	Value: size,
		// },
		Target: &libvirtxml.StorageVolumeTarget{
			Format: &libvirtxml.StorageVolumeTargetFormat{
				Type: "qcow2",
			},
		},
		BackingStore: &libvirtxml.StorageVolumeBackingStore{
			Path: baseImgPath,
			Format: &libvirtxml.StorageVolumeTargetFormat{
				Type: "qcow2",
			},
		},
	}

	xml, err := vol.Marshal()
	if err != nil {
		return nil, fmt.Errorf("could not marshal clone volume: %s", err)
	}

	sv, err := m.l.StorageVolCreateXML(sp, xml, 0)
	return &sv, err
}

func (m *LibvirtManager) Create(name string, baseImage string) error {
	meta, err := cloudinit.GetMetaData(name)
	if err != nil {
		return err
	}

	user, err := cloudinit.GetUserData(name)
	if err != nil {
		return err
	}

	data, err := cloudinit.CreateIso(meta, user, "")
	if err != nil {
		return err
	}
	vol, err := m.createIsoVolume(fmt.Sprintf("config_%s", name), uint64(len(data)), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create config iso: %s", err)
	}
	return m.createVm(name, baseImage, *vol)
}

// Remove removes the domain and its volumes
func (m *LibvirtManager) Remove(name string) error {
	err := m.removeDomain(name)
	if err != nil {
		return err
	}
	err = m.removeVolume(name)
	if err != nil {
		return err
	}
	err = m.removeVolume(fmt.Sprintf("config_%s", name))
	if err != nil {
		return err
	}
	return nil
}
