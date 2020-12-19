package dir

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dvob/vu/internal/image"
)

type Service struct {
	dir string
}

func New(baseDir string) *Service {
	return &Service{
		dir: baseDir,
	}
}

func (s *Service) Add(name string, img io.Reader) (*image.Image, error) {
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

func (s *Service) List() ([]image.Image, error) {
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

func (s *Service) Remove(name string) error {
	return os.Remove(filepath.Join(s.dir, name))
}
