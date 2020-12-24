package internal

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/dvob/vu/internal/cloudinit"
	"github.com/dvob/vu/internal/image"
	"github.com/dvob/vu/internal/vm"
)

type Manager struct {
	ConfigImagePool string
	BaseImagePool   string
	VMImagePool     string
	Image           image.Manager
	VM              vm.Manager
}

func (m *Manager) Create(name, baseImageName string, vmConfig *vm.Config, ciConfig *cloudinit.Config) error {
	baseImage, err := m.Image.Get(m.BaseImagePool, baseImageName)
	if err != nil {
		return err
	}

	image, err := m.Image.Clone(baseImage.ID, m.VMImagePool, name, vmConfig.DiskSize)
	if err != nil {
		return fmt.Errorf("failed to clone image '%s': %w", baseImage, err)
	}
	vmConfig.Image = image.ID

	isoConfig, err := ciConfig.ISO()
	if err != nil {
		return fmt.Errorf("failed to create config ISO: %w", err)
	}

	var reader io.ReadCloser
	reader = ioutil.NopCloser(bytes.NewBuffer(isoConfig))

	isoImage, err := m.Image.Create(m.ConfigImagePool, name, reader)
	if err != nil {
		// try to cleanup cloned base image
		_ = m.Image.Remove(image.ID)
		return fmt.Errorf("failed to store config ISO: %w", err)
	}
	vmConfig.ISO = isoImage.ID

	err = m.VM.Create(name, vmConfig)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) Remove(name string) error {
	state, err := m.VM.Get(name)
	if err != nil {
		return err
	}

	for _, imageID := range state.Images {
		err := m.Image.Remove(imageID)
		if err != nil {
			return err
		}
	}

	return m.VM.Remove(name)
}
