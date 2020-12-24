package libvirt

import (
	"bufio"
	"fmt"
	"io"
	"strings"

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
				// add as read-only since qcow2 base images should not be edited
				Mode: "0444",
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

func (m *Manager) Clone(baseImageName string, name string, newSize uint64) (*image.Image, error) {
	sp, err := m.createOrGetPool()
	if err != nil {
		return nil, fmt.Errorf("faild to get storage pool: %s", err)
	}

	baseImage, err := m.Get(baseImageName)
	if err != nil {
		return nil, fmt.Errorf("failed to get base image %s: %w", baseImageName, err)
	}

	// TODO add owner and group
	vol := &libvirtxml.StorageVolume{
		Name: name,
		Target: &libvirtxml.StorageVolumeTarget{
			Format: &libvirtxml.StorageVolumeTargetFormat{
				// TODO: should not use fix format here
				Type: "qcow2",
			},
		},
		BackingStore: &libvirtxml.StorageVolumeBackingStore{
			Path: baseImage.Location,
			Format: &libvirtxml.StorageVolumeTargetFormat{
				// TODO: should not use fix format here
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
	if err != nil {
		return nil, fmt.Errorf("failed to clone image: %w", err)
	}
	return m.Get(sv.Name)
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

	// TODO change permissions to libvirt user
	// err = os.MkdirAll(m.path, 0777)
	// if err != nil {
	// 	return nil, err
	// }
	pool := libvirtxml.StoragePool{
		Type: "dir",
		Name: m.pool,
		Target: &libvirtxml.StoragePoolTarget{
			Path: m.path,
		},
	}

	xml, err := pool.Marshal()
	if err != nil {
		return nil, err
	}
	sp, err = m.StoragePoolCreateXML(xml, libvirt.StoragePoolCreateWithBuild)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage pool: %w", err)
	}
	// err = m.StoragePoolBuild(sp, 0)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to build storage pool: %w", err)
	// }
	// err = m.StoragePoolCreate(sp, 0)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to build storage pool: %w", err)
	// }
	return &sp, nil
}
