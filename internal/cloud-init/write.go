package cloudinit

import (
	"bytes"
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

// CmdReader implements a io.Reader which return errors of the command execution. Stderr is redirected to
type CmdReader struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

func NewCmdReader(cmd *exec.Cmd) (*CmdReader, error) {
	cmd.Stderr = os.Stderr
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	return &CmdReader{
		cmd:    cmd,
		stdout: reader,
	}, nil
}

// Read reads on stdout. If io.EOF is returned it checks if the command ended
// successfully and if not it returns that error insted of io.EOF
func (cr *CmdReader) Read(p []byte) (n int, err error) {
	n, err = cr.stdout.Read(p)
	if err == io.EOF {
		err1 := cr.cmd.Wait()
		if err1 != nil {
			return n, err1
		}
	}
	return
}

func (cr *CmdReader) Close() error {
	return cr.stdout.Close()
}

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

	r, err := NewCmdReader(cmd)
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
		network, err = c.NetworkConfig.String()
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

	err = c.WriteToDir(tmp)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	isoReader, err := CreateISOFromDir(tmp)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(buf, isoReader)
	if err != nil {
		return nil, err
	}
	err = os.RemoveAll(tmp)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(buf), nil
}
