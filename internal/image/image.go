package image

import "io"

type Service interface {
	Create(name string, image io.Reader) (Image, error)
	List() ([]Image, error)
	Remove(name string) error
}

type Image struct {
	Name     string
	Location string
}
