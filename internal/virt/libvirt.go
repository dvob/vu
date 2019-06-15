package virt

import (
	"fmt"
	"net"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/dsbrng25b/cis/internal/cloud-init"
	"github.com/libvirt/libvirt-go-xml"
)

const CONFIG_VOL_PREFIX = "config_"
const BASE_IMAGE_PREFIX = "cis_base_"

type Manager interface {
	CreateBaseImage(name string, src string) error
	Create(name string, baseImage string) error
	Remove(name string) error
}

type LibvirtManager struct {
	l       *libvirt.Libvirt
	pool    string
	network string
}

type VMConfig struct {
	Memory          uint
	VCPU            int
	Network         string
	BaseImageVolume string
	ConfigVolume    string
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

func NewDefaultVMConfig(name, baseImage string) *VMConfig {
	return &VMConfig{
		Memory:          1024000000, //1024MB
		VCPU:            1,
		Network:         "default",
		BaseImageVolume: baseImage,
		ConfigVolume:    CONFIG_VOL_PREFIX + name,
	}
}

func (m *LibvirtManager) Create(name string, vmCfg *VMConfig, cloudCfg *cloudinit.Config) error {
	_, err := m.cloneBaseImage(name, vmCfg.BaseImageVolume)
	if err != nil {
		return err
	}

	_, err = m.createConfigVolume(vmCfg.ConfigVolume, cloudCfg)
	if err != nil {
		return err
	}

	return m.createVM(name, vmCfg)
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
	err = m.removeVolume(fmt.Sprintf(CONFIG_VOL_PREFIX, name))
	if err != nil {
		return err
	}
	return nil
}

func (m *LibvirtManager) Start(name string) error {
	fmt.Println("not implemented yet")
	return nil
}

func (m *LibvirtManager) Stop(name string) error {
	fmt.Println("not implemented yet")
	return nil
}

func (m *LibvirtManager) List() ([]string, error) {
	fmt.Println("not implemented yet")
	return []string{}, nil
}

func (m *LibvirtManager) createVM(name string, cfg *VMConfig) error {
	baseImage, err := m.GetVolume(cfg.BaseImageVolume)
	if err != nil {
		return fmt.Errorf("failed to clone %s: %s", cfg.BaseImageVolume, err)
	}

	path, err := m.l.StorageVolGetPath(*baseImage)
	if err != nil {
		return err
	}
	configVol, err := m.GetVolume(cfg.ConfigVolume)
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
