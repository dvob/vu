package libvirt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/digitalocean/go-libvirt"
	"github.com/dvob/vu/internal/image"
	libvirtxml "github.com/libvirt/libvirt-go-xml"
)

var _ image.Manager = &Manager{}

type Manager struct {
	path string
	pool string
	*libvirt.Libvirt
}

func New(pool, path string, libvirt *libvirt.Libvirt) *Manager {
	return &Manager{
		path,
		pool,
		libvirt,
	}
}

func (m *Manager) Create(name string, img io.ReadCloser) (*image.Image, error) {
	vol := &libvirtxml.StorageVolume{
		Name: name,
		Capacity: &libvirtxml.StorageVolumeSize{
			Value: 0,
		},
		Target: &libvirtxml.StorageVolumeTarget{
			Permissions: &libvirtxml.StorageVolumeTargetPermissions{
				Owner: strconv.Itoa(syscall.Getuid()),
				Group: strconv.Itoa(syscall.Getgid()),
				// add as read-only since qcow2 base images should not be edited
				Mode: "0400",
			},
		},
	}

	xml, err := vol.Marshal()
	if err != nil {
		return nil, err
	}

	sp, err := m.createOrGetPool()
	if err != nil {
		return nil, fmt.Errorf("faild to get storage pool: %w", err)
	}

	sv, err := m.StorageVolCreateXML(*sp, xml, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	reader := bufio.NewReader(img)

	err = m.StorageVolUpload(sv, reader, 0, 0, 0)
	if err != nil {
		// try undo
		_ = m.StorageVolDelete(sv, 0)
		return nil, fmt.Errorf("failed to upload content: %w", err)
	}

	location, err := m.StorageVolGetPath(sv)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume location: %w", err)
	}
	return &image.Image{
		Name:     name,
		Location: location,
	}, nil
}

func (m *Manager) List() ([]image.Image, error) {
	sp, err := m.StoragePoolLookupByName(m.pool)
	if err != nil {
		return nil, fmt.Errorf("faild to get storage pool: %s", err)
	}

	vols, _, err := m.StoragePoolListAllVolumes(sp, 1, 0)
	if err != nil {
		return nil, err
	}

	images := []image.Image{}
	for _, vol := range vols {
		location, err := m.StorageVolGetPath(vol)
		if err != nil {
			return nil, err
		}
		images = append(images, image.Image{
			Name:     vol.Name,
			Location: location,
		})
	}
	return images, nil
}

func (m *Manager) Remove(name string) error {
	sp, err := m.StoragePoolLookupByName(m.pool)
	if err != nil {
		return fmt.Errorf("faild to get storage pool: %s", err)
	}
	sv, err := m.StorageVolLookupByName(sp, name)
	if err != nil {
		return err
	}
	return m.StorageVolDelete(sv, 0)
}

func (m *Manager) Get(name string) (*image.Image, error) {
	sp, err := m.StoragePoolLookupByName(m.pool)
	if err != nil {
		return nil, fmt.Errorf("faild to get storage pool: %w", err)
	}

	sv, err := m.StorageVolLookupByName(sp, name)
	if err != nil {
		return nil, err
	}

	location, err := m.StorageVolGetPath(sv)
	if err != nil {
		return nil, err
	}
	return &image.Image{
		Name:     name,
		Location: location,
	}, nil
}

func (m *Manager) Clone(name string, baseImageLocation string, newSize uint64) (*libvirt.StorageVol, error) {
	sp, err := m.createOrGetPool()
	if err != nil {
		return nil, fmt.Errorf("faild to get storage pool: %s", err)
	}
	vol := &libvirtxml.StorageVolume{
		Name: name,
		Target: &libvirtxml.StorageVolumeTarget{
			Format: &libvirtxml.StorageVolumeTargetFormat{
				Type: "qcow2",
			},
		},
		BackingStore: &libvirtxml.StorageVolumeBackingStore{
			Path: baseImageLocation,
			Format: &libvirtxml.StorageVolumeTargetFormat{
				Type: "qcow2",
			},
		},
	}

	if newSize != 0 {
		vol.Capacity = &libvirtxml.StorageVolumeSize{
			Value: newSize,
			Unit:  "b",
		}
	}

	xml, err := vol.Marshal()
	if err != nil {
		return nil, err
	}

	sv, err := m.StorageVolCreateXML(*sp, xml, 0)
	return &sv, err
}

func (m *Manager) createOrGetPool() (*libvirt.StoragePool, error) {
	sp, err := m.StoragePoolLookupByName(m.pool)
	if err == nil {
		return &sp, nil
	}

	// TODO: seems that the underlaying errors of libvirt are not exported
	if !strings.Contains(err.Error(), "Storage pool not found") {
		return nil, err
	}

	err = os.MkdirAll(m.path, 0750)
	if err != nil {
		return nil, err
	}
	pool := libvirtxml.StoragePool{
		Type: "dir",
		Name: m.pool,
		Target: &libvirtxml.StoragePoolTarget{
			Path: m.path,
			Permissions: &libvirtxml.StoragePoolTargetPermissions{
				Owner: strconv.Itoa(syscall.Getuid()),
				Group: strconv.Itoa(syscall.Getgid()),
			},
		},
	}

	xml, err := pool.Marshal()
	if err != nil {
		return nil, err
	}
	fmt.Println(string(xml))
	sp, err = m.StoragePoolCreateXML(xml, 0)
	return &sp, err
}
