package dir

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dvob/vu/internal/image"
)

type Manager struct {
	dir string
}

func New(baseDir string) *Manager {
	return &Manager{
		dir: baseDir,
	}
}

func (s *Manager) Create(pool, name string, img io.ReadCloser) (*image.Image, error) {
	dirPath := filepath.Join(s.dir, pool)
	err := os.MkdirAll(dirPath, 0750)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(img)
	targetFile := filepath.Join(dirPath, name)

	file, err := os.Create(targetFile)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		return nil, err
	}

	return &image.Image{
		ID:   targetFile,
		Name: name,
	}, nil

}

func (s *Manager) List(pool string) ([]image.Image, error) {
	dirPath := filepath.Join(s.dir, pool)
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	images := []image.Image{}
	for _, file := range files {
		absPath, err := filepath.Abs(filepath.Join(dirPath, file.Name()))
		if err != nil {
			return nil, err
		}
		images = append(images, image.Image{
			ID:   absPath,
			Name: file.Name(),
		})
	}
	return images, nil
}

func (s *Manager) Remove(ID string) error {
	return os.Remove(ID)
}
