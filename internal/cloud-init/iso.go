package cloudinit

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"io/ioutil"
	"path/filepath"
)

var isoCmd = "mkisofs"
var isoTmpPrefix = "cis_iso"
var maxIsoSize = 100 * 1000 * 1000
var metaFileName = "meta-data"
var userFileName = "user-data"
var networkFileName = "network-config"

func CreateIso(meta, user, network string) ([]byte, error) {
	tmp, err := ioutil.TempDir("", isoTmpPrefix)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	err = ioutil.WriteFile(filepath.Join(tmp, metaFileName), []byte(meta), 0644)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(filepath.Join(tmp, userFileName), []byte(user), 0644)
	if err != nil {
		return nil, err
	}

	if network != "" {
		err = ioutil.WriteFile(filepath.Join(tmp, networkFileName), []byte(network), 0644)
		if err != nil {
			return nil, err
		}
	}

	cmdArgs := []string{
		"-volid",
		"cidata",
		"-joliet",
		"-rock",
		filepath.Join(tmp, userFileName),
		filepath.Join(tmp, metaFileName),
	}

	if network != "" {
		cmdArgs = append(cmdArgs, filepath.Join(tmp, networkFileName))
	}

	cmd := exec.Command(isoCmd, cmdArgs...)

	var reader io.Reader
	reader, err = cmd.StdoutPipe()
	reader = io.LimitReader(reader, int64(maxIsoSize + 1))

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

