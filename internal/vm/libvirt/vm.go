package libvirt

import (
	"fmt"
	"strconv"

	"github.com/digitalocean/go-libvirt"
	"github.com/dvob/vu/internal/vm"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

const (
	UnusedFlag uint32 = 0
)

var _ vm.Manager = &Manager{}

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
			fmt.Printf("    name=%s, hwaddr=%s, Addrs=%v", iface.Name, iface.Hwaddr, iface.Addrs)
		}
	}
	return nil
}

func (m *Manager) List() ([]vm.VM, error) {
	// TODO: not sure why first paramater has to be 1
	domains, _, err := m.ConnectListAllDomains(1, 0)
	if err != nil {
		return nil, err
	}
	vms := []vm.VM{}
	for _, dom := range domains {
		vm, err := m.get(dom)
		if err != nil {
			return nil, err
		}
		vms = append(vms, *vm)
	}
	return vms, nil
}

func (m *Manager) Get(name string) (*vm.VM, error) {
	dom, err := m.DomainLookupByName(name)
	if err != nil {
		return nil, err
	}
	return m.get(dom)
}

func (m *Manager) get(dom libvirt.Domain) (*vm.VM, error) {
	state := &vm.VM{
		Name: dom.Name,
	}

	// get disks
	xml, err := m.DomainGetXMLDesc(dom, 0)
	if err != nil {
		return nil, err
	}
	vmDef := &libvirtxml.Domain{}
	err = vmDef.Unmarshal(xml)
	if err != nil {
		return nil, err
	}
	state.Images = getDisksFromDomain(vmDef)

	// get IP
	state.IPAddress = m.getIP(dom)

	// state
	domState, _, err := m.DomainGetState(dom, UnusedFlag)
	if err != nil {
		return nil, err
	}
	state.State = stateToString(libvirt.DomainState(domState))
	return state, nil
}

func stateToString(state libvirt.DomainState) string {
	switch state {
	case libvirt.DomainRunning:
		return "running"
	case libvirt.DomainBlocked:
		return "blocked"
	case libvirt.DomainPaused:
		return "paused"
	case libvirt.DomainShutdown:
		return "shutdown"
	case libvirt.DomainShutoff:
		return "shutoff"
	case libvirt.DomainCrashed:
		return "crashed"
	case libvirt.DomainPmsuspended:
		return "pmsuspended"
	default:
		return strconv.Itoa(int(state))
	}
}

func (m *Manager) getIP(dom libvirt.Domain) string {
	// dhcp leases
	ifs, err := m.DomainInterfaceAddresses(dom, 0, UnusedFlag)
	if err == nil && len(ifs) > 0 {
		ip := getFirstIP(ifs)
		if ip != "" {
			return ip
		}
	}

	// arp cache
	ifs, err = m.DomainInterfaceAddresses(dom, 2, UnusedFlag)
	if err == nil && len(ifs) > 0 {
		ip := getFirstIP(ifs)
		if ip != "" {
			return ip
		}
	}

	return "n/a"
}

func getFirstIP(ifaces []libvirt.DomainInterface) string {
	for _, iface := range ifaces {
		if len(iface.Addrs) > 0 {
			return iface.Addrs[0].Addr
		}
	}
	return ""
}

func getDisksFromDomain(dom *libvirtxml.Domain) []string {
	disks := []string{}
	if dom.Devices == nil {
		return nil
	}
	for _, disk := range dom.Devices.Disks {
		if disk.Source == nil {
			continue
		}
		disks = append(disks, disk.Source.File.File)
	}
	return disks
}

func (m *Manager) Create(name string, cfg *vm.Config) error {
	domain := &libvirtxml.Domain{
		Name:        name,
		Type:        "kvm",
		Description: "created by vu",
		Memory: &libvirtxml.DomainMemory{
			Value: uint(cfg.Memory),
			Unit:  "b",
		},
		VCPU: &libvirtxml.DomainVCPU{
			Value: cfg.CPUCount,
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
				{
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
							File: cfg.Image,
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
							File: cfg.ISO,
						},
					},
				},
			},
			Interfaces: []libvirtxml.DomainInterface{
				{
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
				{},
			},
			Videos: []libvirtxml.DomainVideo{
				{
					Model: libvirtxml.DomainVideoModel{
						Type: "vga",
					},
				},
			},
		},
	}

	xml, err := domain.Marshal()
	if err != nil {
		return nil
	}

	dom, err := m.DomainDefineXML(xml)
	if err != nil {
		return fmt.Errorf("failed to define domain: %w", err)
	}

	err = m.DomainCreate(dom)
	if err != nil {
		return fmt.Errorf("failed to create domain: %w", err)
	}
	return err
}

// Remove removes the domain and its volumes
func (m *Manager) Remove(name string) error {
	dom, err := m.DomainLookupByName(name)
	if err != nil {
		return err
	}

	stateInt, _, err := m.DomainGetState(dom, UnusedFlag)
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
