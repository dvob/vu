package virt

import (
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/dsbrng25b/cis/internal/cloud-init"
	"github.com/libvirt/libvirt-go-xml"
)

const configVolPrefix = "config_"
const baseImagePrefix = "cis_base_"

type LibvirtManager struct {
	l    *libvirt.Libvirt
	pool string
}

type VMConfig struct {
	Memory          uint
	VCPU            int
	Network         string
	BaseImageVolume string
	DiskSize        uint64
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
	}, nil
}

func (m *LibvirtManager) Create(name string, vmCfg *VMConfig, cloudCfg *cloudinit.Config) error {
	mainVol, err := m.cloneBaseImage(name, vmCfg.BaseImageVolume, vmCfg.DiskSize)
	if err != nil {
		return err
	}

	configVol, err := m.createConfigVolume(configVolPrefix+name, cloudCfg)
	if err != nil {
		m.l.StorageVolDelete(*mainVol, 0)
		return err
	}

	err = m.createVM(name, vmCfg)
	if err != nil {
		m.l.StorageVolDelete(*mainVol, 0)
		m.l.StorageVolDelete(*configVol, 0)
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
	err = m.removeVolume(name)
	if err != nil {
		return err
	}
	err = m.removeVolume(configVolPrefix + name)
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

func (m *LibvirtManager) List() ([]string, error) {
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

func (m *LibvirtManager) createVM(name string, cfg *VMConfig) error {
	image, err := m.GetVolume(name)
	if err != nil {
		return err
	}

	path, err := m.l.StorageVolGetPath(*image)
	if err != nil {
		return err
	}
	configVol, err := m.GetVolume(configVolPrefix + name)
	if err != nil {
		return err
	}
	configPath, err := m.l.StorageVolGetPath(*configVol)
	if err != nil {
		return err
	}

	domain := &libvirtxml.Domain{
		Name: name,
		Type: "kvm",
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
						Dev: "hda",
						Bus: "ide",
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
							Network: cfg.Network,
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
