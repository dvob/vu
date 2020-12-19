package image

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/cheggaaa/pb.v1"
)

func AddFromURL(service Manager, name, src string, progress io.Writer) (*Image, error) {
	u, err := url.Parse(src)

	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %s", err)
	}

	var (
		reader io.ReadCloser
		size   uint64
	)
	if u.Scheme == "file" {
		file, err := os.Open(u.Path)
		if err != nil {
			return nil, err
		}

		fileinfo, err := file.Stat()
		if err != nil {
			return nil, err
		}

		size = uint64(fileinfo.Size())
		reader = file
	} else if u.Scheme == "http" || u.Scheme == "https" {
		resp, err := http.Get(u.String())
		if err != nil {
			return nil, err
		}

		if resp.StatusCode > 399 {
			return nil, fmt.Errorf("http status %d returned", resp.StatusCode)
		}

		size = uint64(resp.ContentLength)
		reader = resp.Body

	} else {
		return nil, fmt.Errorf("unkown schema '%s'", u.Scheme)
	}

	var bar *pb.ProgressBar
	if progress != nil {
		bar = pb.New(int(size)).SetUnits(pb.U_BYTES)
		bar.Start()
		reader = bar.NewProxyReader(reader)
	}
	if name == "" {
		name = filepath.Base(u.Path)
	}
	image, err := service.Create(name, reader)
	if err != nil {
		return nil, err
	}
	if progress != nil {
		bar.Finish()
	}
	return image, reader.Close()
}
