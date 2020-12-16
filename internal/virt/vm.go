package virt

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/libvirt/libvirt-go-xml"
)

const descriptionPrefix = "created by vu"

type LibvirtManager struct {
	l    *libvirt.Libvirt
	pool string
}

type VMConfig struct {
	Name            string
	Memory          uint
	VCPU            uint
	Network         string
	BaseImageVolume string
	ISOImageVolume  string
	DiskSize        uint64
}

func NewLibvirtManager(pool string, uri string) (*LibvirtManager, error) {
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

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return &LibvirtManager{
		l,
		pool,
	}, nil
}

// Close closes the connection to libvirt
func (m *LibvirtManager) Close() error {
	return m.l.Disconnect()
}

// Create creates a VM from vmCfg
// First it clones the base image configured in vmCfg to the name of the VM and
// then it tries to create the VM itself. If the creation of the VM fails it
// tries to remove the cloned image as well.
func (m *LibvirtManager) Create(vmCfg *VMConfig) error {
	mainVol, err := m.cloneBaseImage(vmCfg.Name, vmCfg.BaseImageVolume, vmCfg.DiskSize)
	if err != nil {
		return err
	}

	err = m.createVM(vmCfg)
	if err != nil {
		m.l.StorageVolDelete(*mainVol, 0) //nolint:errcheck
		return err
	}
	return nil

}

// Remove removes the domain and its volumes
func (m *LibvirtManager) Remove(name string) error {
	err := m.removeDomain(name)
	if err != nil {
		return err
	}
	err = m.RemoveVolume(name)
	if err != nil {
		return err
	}
	return nil
}

func (m *LibvirtManager) Start(name string) error {
	dom, err := m.l.DomainLookupByName(name)
	if err != nil {
		return err
	}

	return m.l.DomainCreate(dom)
}

func (m *LibvirtManager) Shutdown(name string, force bool) error {
	dom, err := m.l.DomainLookupByName(name)
	if err != nil {
		return err
	}

	if force {
		return m.l.DomainDestroy(dom)
	}
	return m.l.DomainShutdown(dom)
}

func (m *LibvirtManager) ListAll() ([]string, error) {
	var domNames = []string{}
	// TODO: not sure why first paramater has to be 1
	domains, _, err := m.l.ConnectListAllDomains(1, 0)
	if err != nil {
		return nil, err
	}
	for _, dom := range domains {
		domNames = append(domNames, dom.Name)
	}
	sort.Strings(domNames)
	return domNames, nil
}

func (m *LibvirtManager) List() ([]string, error) {
	var domNames = []string{}
	// TODO: not sure why first paramater has to be 1
	domains, _, err := m.l.ConnectListAllDomains(1, 0)
	if err != nil {
		return nil, err
	}
	for _, dom := range domains {
		description, err := m.l.DomainGetMetadata(dom, 0, nil, 0)
		if err != nil {
			//TODO: explicitly check for metadata not available error
			continue
		}
		if strings.HasPrefix(description, descriptionPrefix) {
			domNames = append(domNames, dom.Name)
		}
	}
	sort.Strings(domNames)
	return domNames, nil
}

func (m *LibvirtManager) createVM(cfg *VMConfig) error {
	image, err := m.GetVolume(cfg.Name)
	if err != nil {
		return err
	}

	path, err := m.l.StorageVolGetPath(*image)
	if err != nil {
		return err
	}

	configVol, err := m.GetVolume(cfg.ISOImageVolume)
	if err != nil {
		return err
	}

	configPath, err := m.l.StorageVolGetPath(*configVol)
	if err != nil {
		return err
	}

	domain := &libvirtxml.Domain{
		Name:        cfg.Name,
		Type:        "kvm",
		Description: descriptionPrefix,
		Memory: &libvirtxml.DomainMemory{
			Value: cfg.Memory,
			Unit:  "b", //bytes
		},
		VCPU: &libvirtxml.DomainVCPU{
			Value: cfg.VCPU,
		},
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Type: "hvm",
			},
		},
		Features: &libvirtxml.DomainFeatureList{
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
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
						Dev: "vda",
						Bus: "virtio",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: path,
						},
					},
				}, {
					Device: "cdrom",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "raw",
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "vdb",
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
							Network: cfg.Network,
						},
					},
					Model: &libvirtxml.DomainInterfaceModel{
						Type: "virtio",
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
