package interfaces

import (
	"github.com/Sentimentron/repositron/models"
	"io"
)

type OpaqueImageHandle int

// ImageSpecification specifies the initial contents of an image
type ImageSpecification struct {
	DockerFile string `json:"dockerFileString"`
	ImageName  string `json:"imageName"`
	// This is a Repositron link to a tar file which contains
	// the code/data loaded inside the image.
	Contents []models.Blob `json:"initialContents"`
}

// OutputStream is a way of handling Stderr, Stdout streams
type OutputStream interface {
	Stderr() io.Writer
	Stdout() io.Writer

	StderrReader() io.Reader
	StdoutReader() io.Reader
}

// ImageLibrary is a way of building and managing templated containers.
type ImageLibrary interface {

	// Returns true if an image has been previously built with the provided name.
	CheckImageBuilt(name string) (bool, error)

	// AcquireImage updates the most recently used time of the provided image,
	// prevents deletion until ReleaseImage is called.
	AcquireImage(name string) (OpaqueImageHandle, error)

	// ReleaseImage marks the image as unused. Must pass the handle obtained via
	// AcquireImage earlier.
	ReleaseImage(handle OpaqueImageHandle) error

	// DeleteImage cleans up a container immediately if it's no longer used.
	DeleteImage(name string) error

	// BuildImage creates an image according to the given specification.
	BuildImage(name string, spec *ImageSpecification, monitor OutputStream) error

	// GetImages returns a list of all the images which have been built so far.
	GetImages() ([]string, error)
}

// DockerCommandRunner is an interface over docker, provided for testing.
type DockerCommandRunner interface {
	ListImages() (string, error)
}