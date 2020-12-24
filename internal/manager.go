package internal

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/dvob/vu/internal/cloudinit"
	"github.com/dvob/vu/internal/image"
	"github.com/dvob/vu/internal/vm"
)

type Manager struct {
	ConfigImage image.Manager
	BaseImage   image.Manager
	VM          vm.Manager
}

func (m *Manager) Create(name, baseImage string, vmConfig *vm.Config, ciConfig *cloudinit.Config) error {
	image, err := m.BaseImage.Clone(baseImage, name, vmConfig.DiskSize)
	if err != nil {
		return fmt.Errorf("failed to clone image '%s': %w", baseImage, err)
	}
	vmConfig.Image = image.Location

	isoConfig, err := ciConfig.ISO()
	if err != nil {
		return fmt.Errorf("failed to create config ISO: %w", err)
	}

	var reader io.ReadCloser
	reader = ioutil.NopCloser(bytes.NewBuffer(isoConfig))

	isoImage, err := m.ConfigImage.Create(name, reader)
	if err != nil {
		_ = m.BaseImage.Remove(name)
		return fmt.Errorf("failed to store config ISO: %w", err)
	}
	vmConfig.ISO = isoImage.Location

	err = m.VM.Create(name, vmConfig)
	if err != nil {
		//_ = m.BaseImage.Remove(name)
		//_ = m.ConfigImage.Remove(name)
		return err
	}
	return nil
}

func (m *Manager) Remove(name string) error {
	state, err := m.VM.Get(name)
	if err != nil {
		return err
	}

	for _, imagePath := range state.Images {
		// TODO: just NO. use proper abstraction here
		var mgr image.Manager
		if strings.Contains(imagePath, "config") {
			mgr = m.ConfigImage
		} else {
			mgr = m.BaseImage
		}

		name := filepath.Base(imagePath)
		err := mgr.Remove(name)
		if err != nil {
			return err
		}
	}

	return m.VM.Remove(name)
}
