package virt

import (
	"fmt"

	"github.com/digitalocean/go-libvirt"
	"github.com/dvob/vu/internal/vm"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

type Manager struct {
	*libvirt.Libvirt
}

func New(libvirt *libvirt.Libvirt) *Manager {
	return &Manager{
		libvirt,
	}
}

func (m *Manager) Start(name string) error {
	dom, err := m.DomainLookupByName(name)
	if err != nil {
		return err
	}

	return m.DomainCreate(dom)
}

func (m *Manager) Shutdown(name string, force bool) error {
	dom, err := m.DomainLookupByName(name)
	if err != nil {
		return err
	}

	if force {
		return m.DomainDestroy(dom)
	}
	return m.DomainShutdown(dom)
}

func (m *Manager) ListDetail() error {
	domains, _, err := m.ConnectListAllDomains(1, 0)
	if err != nil {
		return err
	}
	for _, dom := range domains {
		ifaces, err := m.DomainInterfaceAddresses(dom, 2, 0)
		if err != nil {
			return fmt.Errorf("failed to get domain interface address for %s: %w", dom.Name, err)
		}
		fmt.Println(dom.Name)
		for _, iface := range ifaces {
			fmt.Printf("    %#v\n", iface)
			fmt.Printf("    name=%s, hwaddr=%s, Addrs=%s", iface.Name, iface.Hwaddr, iface.Addrs)
		}
	}
	return nil
}

func (m *Manager) List() ([]string, error) {
	var domNames = []string{}
	// TODO: not sure why first paramater has to be 1
	domains, _, err := m.ConnectListAllDomains(1, 0)
	if err != nil {
		return nil, err
	}
	for _, dom := range domains {
		domNames = append(domNames, dom.Name)
	}
	return domNames, nil
}

func (m *Manager) Create(cfg *vm.Config) error {
	domain := &libvirtxml.Domain{
		Name:        cfg.Name,
		Type:        "kvm",
		Description: "created by vu",
		Memory: &libvirtxml.DomainMemory{
			Value: cfg.Memory,
			Unit:  "b",
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
							File: cfg.BaseImage,
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
							File: cfg.ConfigISO,
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

	dom, err := m.DomainDefineXML(xml)
	if err != nil {
		return err
	}

	err = m.DomainCreate(dom)
	return err
}

// Remove removes the domain and its volumes
func (m *Manager) Remove(name string) error {
	err := m.removeVolume(name)
	if err != nil {
		return err
	}
	err = m.removeDomain(name)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) removeVolume(name string) error {
	fmt.Println("remove volume not yet implemented")
	return nil
}

func (m *Manager) removeDomain(name string) error {
	dom, err := m.DomainLookupByName(name)
	if err != nil {
		return err
	}

	stateInt, _, err := m.DomainGetState(dom, 0)
	state := libvirt.DomainState(stateInt)
	if err != nil {
		return err
	}
	if state == libvirt.DomainRunning || state == libvirt.DomainPaused {
		err = m.DomainDestroy(dom)
		if err != nil {
			return err
		}
	}

	err = m.DomainUndefine(dom)
	if err != nil {
		return err
	}

	return nil
}
