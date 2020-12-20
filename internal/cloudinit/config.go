package cloudinit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// Config represents all configurations for cloud init
type Config struct {
	MetaData      *MetaData
	UserData      *UserData
	NetworkConfig *NetworkConfig
}

var (
	tmpPrefix       = "vu_iso"
	metaFileName    = "meta-data"
	userFileName    = "user-data"
	networkFileName = "network-config"
)

func (c *Config) Merge(c2 *Config) error {
	if c.MetaData == nil {
		c.MetaData = c2.MetaData
	} else {
		err := c.MetaData.Merge(c2.MetaData)
		if err != nil {
			return err
		}
	}
	if c.UserData == nil {
		c.UserData = c2.UserData
	} else {
		err := c.UserData.Merge(c2.UserData)
		if err != nil {
			return err
		}
	}
	if c.NetworkConfig == nil {
		c.NetworkConfig = c2.NetworkConfig
	} else {
		err := c.NetworkConfig.Merge(c2.NetworkConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

// ConfigFromDir reads cloud-init configuration from a directory
func ConfigFromDir(dir string) (*Config, error) {
	_, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}

	c := &Config{}

	err = unmarshalFromFile(filepath.Join(dir, metaFileName), c.MetaData)
	if err != nil {
		return nil, err
	}
	err = unmarshalFromFile(filepath.Join(dir, userFileName), c.UserData)
	if err != nil {
		return nil, err
	}
	err = unmarshalFromFile(filepath.Join(dir, networkFileName), c.NetworkConfig)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ToDir writes the cloud init configuration to a directory
func (c *Config) ToDir(dir string) error {
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return err
	}

	err = marshalToFile(filepath.Join(dir, metaFileName), c.MetaData)
	if err != nil {
		return err
	}

	err = marshalToFile(filepath.Join(dir, userFileName), c.UserData)
	if err != nil {
		return err
	}

	err = marshalToFile(filepath.Join(dir, networkFileName), c.NetworkConfig)
	return err
}

// ISO returns the cloud init configuration as ISO image
func (c *Config) ISO() ([]byte, error) {
	tmp, err := ioutil.TempDir("", tmpPrefix)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	err = c.ToDir(tmp)
	if err != nil {
		return nil, err
	}

	stdErr := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}

	cmdArgs := []string{
		"-volid",
		"cidata",
		"-joliet",
		"-rock",
		tmp,
	}
	cmd := exec.Command("mkisofs", cmdArgs...)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("iso creation failed. stderr: '%s', err: %w", stdErr.String(), err)
	}
	return stdOut.Bytes(), nil
}

func (c *Config) String() (string, error) {
	var (
		meta    []byte
		user    []byte
		network []byte
		err     error
		buf     bytes.Buffer
	)
	if c.MetaData != nil {
		meta, err = c.MetaData.Marshal()
		if err != nil {
			return "", fmt.Errorf("could not render meta data: %s", err)
		}
		buf.WriteString("### meta-data ###\n")
		buf.Write(meta)
		buf.WriteString("\n")
	}

	if c.UserData != nil {
		user, err = c.UserData.Marshal()
		if err != nil {
			return "", fmt.Errorf("could not render user data: %s", err)
		}
		buf.WriteString("### user-data ###\n")
		buf.Write(user)
		buf.WriteString("\n")
	}

	if c.NetworkConfig != nil {
		network, err = c.NetworkConfig.Marshal()
		if err != nil {
			return "", fmt.Errorf("could not render network config: %s", err)
		}
		buf.WriteString("### network-config ###\n")
		buf.Write(network)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}
