package image

import "io"

// Manager is the iterface which describes the image management. With create we can create an image in a certain pool (e.g. config isos, base images, images). The pool is to distingush between various image categories. It allows to retrieve only images of a certain type.
type Manager interface {
	Create(pool, name string, image io.ReadCloser) (*Image, error)
	Clone(baseImageID, targetPool, targetName string, size uint64) (*Image, error)
	List(pool string) ([]Image, error)
	Get(pool, name string) (*Image, error)
	Remove(ID string) error
}

type Image struct {
	// ID is a unique identifier for the image. With this identifier the vm.Manager has to be able to identify the image.
	ID   string
	Name string
}
