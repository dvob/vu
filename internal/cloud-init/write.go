package cloudinit

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

var isoCmd = "mkisofs"
var tmpPrefix = "vu_iso"
var metaFileName = "meta-data"
var userFileName = "user-data"
var networkFileName = "network-config"

// Creates a ISO Image from a directory
func CreateISOFromDir(dir string) (io.ReadCloser, error) {
	cmdArgs := []string{
		"-volid",
		"cidata",
		"-joliet",
		"-rock",
		dir,
	}

	cmd := exec.Command(isoCmd, cmdArgs...)

	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	return r, err
}

// WriteToDir writes the cloud init configuration to a directory
func (c *Config) WriteToDir(dir string) error {
	var (
		meta    string
		user    string
		network string
		err     error
	)

	// render configurations
	meta, err = c.MetaData.String()
	if err != nil {
		return err
	}

	user, err = c.UserData.String()
	if err != nil {
		return err
	}

	if c.NetworkConfig != nil {
		network, err = c.UserData.String()
		if err != nil {
			return err
		}
	}

	// write configurations

	err = ioutil.WriteFile(filepath.Join(dir, metaFileName), []byte(meta), 0644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(dir, userFileName), []byte(user), 0644)
	if err != nil {
		return err
	}

	if c.NetworkConfig != nil {
		err = ioutil.WriteFile(filepath.Join(dir, networkFileName), []byte(network), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateISO returns the cloud init configuration as ISO image
func (c *Config) CreateISO() (io.ReadCloser, error) {
	tmp, err := ioutil.TempDir("", tmpPrefix)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	err = c.WriteToDir(tmp)
	if err != nil {
		return nil, err
	}
	return CreateISOFromDir(tmp)
}
