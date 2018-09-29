package interfaces

import (
	"io"
	"github.com/Sentimentron/functron/models"
	"errors"
)

type OpaqueImageHandle int

// OutputStream is a way of handling Stderr, Stdout streams
type OutputStream interface {
	Stderr() io.Writer
	Stdout() io.Writer

	StderrReader() io.Reader
	StdoutReader() io.Reader
}

var NoMatchingImage = errors.New("No matching image")

// ImageStore provides Functron's memory of what images it's built so far.
type ImageStore interface {

	// PersistImageForBuild persists an input image into the store.
	PersistImageForBuild(image *models.FunctronImage) (*models.FunctronImage, error)

	// RetrieveImageByName returns an image for inspection or further activity
	RetrieveImageByName(name string) (*models.FunctronImage, error)

	// RetrieveImages returns a list of all the images which have been built so far.
	RetrieveImages() ([]string, error)

	// UpdateStatus changes the reported status of the image
	// Updates the provided image argument in-place.
	UpdateStatus(image *models.FunctronImage, newStatus models.ImageStatus) error

	// RetrieveBuildPlan returns a struct which contains the images
	// which need to be cleaned up, built etc.
	RetrieveBuildPlan() (*models.BuildPlan, error)

}

// DockerCommandRunner is an interface over docker, provided for testing.
type DockerCommandRunner interface {
	ListImages() (string, error)
}