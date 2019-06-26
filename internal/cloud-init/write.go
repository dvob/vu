package cloudinit

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

var isoCmd = "mkisofs"
var tmpPrefix = "cis_iso"
var maxIsoSize = 100 * 1000 * 1000
var metaFileName = "meta-data"
var userFileName = "user-data"
var networkFileName = "network-config"

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
func (c *Config) CreateISO() ([]byte, error) {
	tmp, err := ioutil.TempDir("", tmpPrefix)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	err = c.WriteToDir(tmp)
	if err != nil {
		return nil, err
	}

	cmdArgs := []string{
		"-volid",
		"cidata",
		"-joliet",
		"-rock",
		filepath.Join(tmp, userFileName),
		filepath.Join(tmp, metaFileName),
	}

	if c.NetworkConfig != nil {
		cmdArgs = append(cmdArgs, filepath.Join(tmp, networkFileName))
	}

	cmd := exec.Command(isoCmd, cmdArgs...)

	// use LimitReader since later we use ReadAll which could easily lead to out of memory
	var reader io.Reader
	reader, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	reader = io.LimitReader(reader, int64(maxIsoSize+1))

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	isoData, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if len(isoData) > maxIsoSize {
		return nil, fmt.Errorf("configurations to big")
	}

	return isoData, nil
}
