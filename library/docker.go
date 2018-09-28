package library

import (
	"github.com/Sentimentron/functron/interfaces"
	"strings"
	"sync"
	"errors"
	"fmt"
)

type DockerImageLibrary struct {
	runner interfaces.DockerCommandRunner
	refCount map[string]int
	handleMap map[interfaces.OpaqueImageHandle]string
	nextHandle interfaces.OpaqueImageHandle
	lock sync.Mutex
}

func CreateDockerImageLibrary(runner interfaces.DockerCommandRunner) *DockerImageLibrary {
	return &DockerImageLibrary{
		runner,
	make(map[string]int),
	make(map[interfaces.OpaqueImageHandle]string),
	1,
	sync.Mutex{}}
}

func FormatToFunctronImageName(shortName string) string {
	return fmt.Sprintf("functron-%s",shortName)
}

func (d *DockerImageLibrary) CheckImageBuilt(name string) (bool, error) {
	// Get a list of all the available images
	availableImages, err := d.runner.ListImages()
	if err != nil {
		return false, err
	}

	for _, line := range strings.Split(availableImages, "\n") {
		components := strings.Split(line, "\t")
		if len(components) > 0 {
			imageName := strings.TrimSpace(components[0])
			if strings.HasPrefix(imageName, "functron-") {
				shortImageName := strings.Replace(imageName, "functron-","",1)
				if shortImageName == name {
					return true, nil
				}
			}
		}
	}
	return false, err
}

func (d *DockerImageLibrary) AcquireImage(name string) (interfaces.OpaqueImageHandle, error) {

	// Acquire the lock to make sure that nothing can delete images whilst we're doing this.
	d.lock.Lock()
	defer d.lock.Unlock()

	// Check that the image exists
	imageBuiltYet, err := d.CheckImageBuilt(name)
	if !imageBuiltYet {
		return -1, errors.New("image: not built")
	} else if err != nil {
		return -1, err
	}

	// If it does, assign the handle
	if _, ok := d.refCount[name]; !ok {
		d.refCount[name] = 0
	}
	d.refCount[name] += 1

	// Acquire and assign the handle
	handle := d.nextHandle
	d.handleMap[handle] = name
	d.nextHandle += 1

	return interfaces.OpaqueImageHandle(handle), err
}

func (d *DockerImageLibrary) ReleaseImage(handle interfaces.OpaqueImageHandle) error {

	// Acquire the lock so nothing else can happen whilst we're doing this
	d.lock.Lock()
	defer d.lock.Unlock()

	// Check that the handle references something
	if _, ok := d.handleMap[handle]; !ok {
		return errors.New("handle invalid")
	}

	// Dereference the handle
	imageName := d.handleMap[handle]
	if _, ok := d.refCount[imageName]; !ok {
		return errors.New("consistency error")
	}
	d.refCount[imageName] -= 1

	return nil
}

func (d *DockerImageLibrary) DeleteImage(name string) error {
	// Acquire the lock so nothing else can happen whilst we're doing this
	d.lock.Lock()
	defer d.lock.Unlock()

	// Check that the image exists
	imageBuiltYet, err := d.CheckImageBuilt(name)
	if !imageBuiltYet {
		return errors.New("image: not built")
	} else if err != nil {
		return err
	}

	// Check that there's nothing still referencing it
	if count, ok := d.refCount[name]; ok {
		if count != 0 {
			return errors.New("image: still in use")
		}
	}

	// Issue the command to Docker to remove the image
	d.runner.RemoveImage(FormatToFunctronImageName(name))
}

func (d *DockerImageLibrary) BuildImage(name string, spec *interfaces.ImageSpecification, monitor interfaces.OutputStream) {

	// Create a temporary directory

	// Download all Repositron specs to that directory

	// Write the Docker file to the directory

	// Pass the context to the Docker agent

}