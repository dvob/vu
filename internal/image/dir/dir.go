package dir

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dvob/vu/internal/image"
)

var _ image.Manager = &Manager{}

type Manager struct {
	dir string
}

func New(baseDir string) *Manager {
	return &Manager{
		dir: baseDir,
	}
}

func (s *Manager) Create(name string, img io.ReadCloser) (*image.Image, error) {
	err := os.MkdirAll(s.dir, 0750)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(img)
	targetFile := filepath.Join(s.dir, name)

	file, err := os.Create(targetFile)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		return nil, err
	}

	return &image.Image{
		Name:     name,
		Location: targetFile,
	}, nil

}

func (s *Manager) List() ([]image.Image, error) {
	files, err := ioutil.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	images := []image.Image{}
	for _, file := range files {
		images = append(images, image.Image{
			Name:     file.Name(),
			Location: filepath.Join(s.dir, file.Name()),
		})
	}
	return images, nil
}

func (s *Manager) Remove(name string) error {
	return os.Remove(filepath.Join(s.dir, name))
}
