package virt

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"

	"github.com/digitalocean/go-libvirt"
	"github.com/dsbrng25b/cis/internal/cloud-init"
	"github.com/libvirt/libvirt-go-xml"
	"gopkg.in/cheggaaa/pb.v1"
)

func (m *LibvirtManager) ImageList() ([]string, error) {
	var volNames = []string{}

	sp, err := m.l.StoragePoolLookupByName(m.pool)
	if err != nil {
		return nil, fmt.Errorf("faild to get storage pool: %s", err)
	}

	vols, _, err := m.l.StoragePoolListAllVolumes(sp, 1, 0)
	if err != nil {
		return nil, err
	}

	for _, vol := range vols {
		volNames = append(volNames, vol.Name)
	}

	sort.Strings(volNames)
	return volNames, nil
}

// creates a volume and uploads the image from the url src into the volume
func (m *LibvirtManager) CreateBaseImage(name string, src string) error {
	var size uint64
	var stream io.Reader

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

	bar := pb.New(int(size)).SetUnits(pb.U_BYTES)
	bar.Start()
	stream = bar.NewProxyReader(stream)

	_, err = m.createVolume(name, size, stream, "qcow2")
	bar.Finish()
	if err != nil {
		return err
	}
	return nil
}

func (m *LibvirtManager) ImageRemove(name string) error {
	return m.removeVolume(name)
}

func (m *LibvirtManager) GetVolume(name string) (*libvirt.StorageVol, error) {
	sp, err := m.l.StoragePoolLookupByName(m.pool)
	if err != nil {
		return nil, fmt.Errorf("faild to get storage pool: %s", err)
	}
	vol, err := m.l.StorageVolLookupByName(sp, name)
	return &vol, err
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

func (m *LibvirtManager) createIsoVolume(name string, size uint64, stream io.Reader) (*libvirt.StorageVol, error) {
	return m.createVolume(name, size, stream, "iso")
}

func (m *LibvirtManager) createVolume(name string, size uint64, stream io.Reader, kind string) (*libvirt.StorageVol, error) {

	vol := &libvirtxml.StorageVolume{
		Name: name,
		Capacity: &libvirtxml.StorageVolumeSize{
			Value: size,
		},
		Target: &libvirtxml.StorageVolumeTarget{
			Format: &libvirtxml.StorageVolumeTargetFormat{
				Type: kind,
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
		// try undo
		m.l.StorageVolDelete(sv, 0) //nolint:errcheck
		return nil, fmt.Errorf("failed to upload content: %s", err)
	}
	return &sv, nil
}

func (m *LibvirtManager) createConfigVolume(name string, cfg *cloudinit.Config) (*libvirt.StorageVol, error) {
	data, err := cfg.CreateISO()
	if err != nil {
		return nil, err
	}
	vol, err := m.createIsoVolume(name, uint64(len(data)), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create config iso: %s", err)
	}
	return vol, nil
}

// cloneBaseImage creates a clone from baseImage. baseImage has to be a qcow2 image.
// With newSize the size of the new image can be specified. If newSize == 0 the size of the image
// is the same as the base image has.
func (m *LibvirtManager) cloneBaseImage(name string, baseImage string, newSize uint64) (*libvirt.StorageVol, error) {
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

	if newSize != 0 {
		vol.Capacity = &libvirtxml.StorageVolumeSize{
			Value: newSize,
			Unit:  "b",
		}
	}

	xml, err := vol.Marshal()
	if err != nil {
		return nil, fmt.Errorf("could not marshal clone volume: %s", err)
	}

	sv, err := m.l.StorageVolCreateXML(sp, xml, 0)
	return &sv, err
}
