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

// NewDefaultConfig returns a minimal default config based on name, user and
// ssh public key
func NewDefaultConfig(name, user, sshPubKey string) *Config {
	return &Config{
		MetaData: &MetaData{
			Hostname:   name,
			InstanceID: name,
		},
		UserData: &UserData{
			Users: []User{
				{
					Name: user,
					Sudo: "ALL=(ALL) NOPASSWD:ALL",
					SSHAuthorizedKeys: []string{
						sshPubKey,
					},
				},
			},
		},
	}
}

// Merge merges configuration c2 into Config. Configurations in c2 overwrite
// configurations in Config.
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

// ConfigFromDir reads cloud-init configuration from directories. If multiple
// directories are passed, configurations of later directories overwrite
// configurations of previous directories
func ConfigFromDir(dirs ...string) (*Config, error) {
	config := &Config{}

	for _, dir := range dirs {
		c, err := configFromDir(dir)
		if err != nil {
			return nil, err
		}
		err = config.Merge(c)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

// configFromDir reads cloud-init configuration from a directory
func configFromDir(dir string) (*Config, error) {
	_, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}

	metaFile := filepath.Join(dir, metaFileName)
	userFile := filepath.Join(dir, userFileName)
	networkFile := filepath.Join(dir, networkFileName)

	c := &Config{}

	data, err := ioutil.ReadFile(metaFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		c.MetaData = &MetaData{}
		err := c.MetaData.Unmarshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to read '%s': %w", metaFile, err)
		}
	}

	data, err = ioutil.ReadFile(userFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		c.UserData = &UserData{}
		err := c.UserData.Unmarshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to read '%s': %w", userFile, err)
		}
	}

	data, err = ioutil.ReadFile(networkFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		c.NetworkConfig = &NetworkConfig{}
		err := c.NetworkConfig.Unmarshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to read '%s': %w", networkFile, err)
		}
	}

	return c, nil
}

// ToDir writes the cloud init configuration to a directory
func (c *Config) ToDir(dir string) error {
	err := os.MkdirAll(dir, 0o750)
	if err != nil {
		return err
	}

	if c.MetaData != nil {
		err = marshalToFile(filepath.Join(dir, metaFileName), c.MetaData)
		if err != nil {
			return err
		}
	}

	if c.UserData != nil {
		err = marshalToFile(filepath.Join(dir, userFileName), c.UserData)
		if err != nil {
			return err
		}
	}

	if c.NetworkConfig != nil {
		err = marshalToFile(filepath.Join(dir, networkFileName), c.NetworkConfig)
		return err
	}
	return nil
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

// String returns a string representation of all cloud-init configs
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
