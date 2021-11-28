package cloudinit

import (
	"bytes"

	"github.com/kdomanski/iso9660"
)

// ISO returns the cloud init configuration as ISO image
func (c *Config) ISO() ([]byte, error) {
	iw, err := iso9660.NewWriter()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = iw.Cleanup()
	}()
	if c.MetaData != nil {
		err = marshalToIso(iw, metaFileName, c.MetaData)
		if err != nil {
			return nil, err
		}
	}

	if c.UserData != nil {
		err = marshalToIso(iw, userFileName, c.UserData)
		if err != nil {
			return nil, err
		}
	}

	if c.NetworkConfig != nil {
		err = marshalToIso(iw, networkFileName, c.NetworkConfig)
		return nil, err
	}
	isoImage := &bytes.Buffer{}
	err = iw.WriteTo(isoImage, "cidata")
	return isoImage.Bytes(), err
}

func marshalToIso(iw *iso9660.ImageWriter, file string, m Marshaler) error {
	if m == nil {
		return nil
	}
	data, err := m.Marshal()
	if err != nil {
		return err
	}
	return iw.AddFile(bytes.NewBuffer(data), file)
}
