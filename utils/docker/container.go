package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
)

// Container represents a docker container and the operations that can be run on it
type Container struct {
	id string
}

// ContainerFromId returns a new Container instance from the given container id
func ContainerFromId(id string) *Container {
	return &Container{id:id}
}

// StartContainer starts a container with the given image, entrypoint, cmd and host mounts
// it wraps the container ID in a Container instance and returns it
func StartContainer(image string, entrypoint []string, cmd []string, mounts []mount.Mount) (*Container, error) {
	ctx := context.Background()
	cli := GetClient()

	logrus.Debug("looking up image ", image)

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}

	logrus.Debug("pulling image ", image)
	io.Copy(ioutil.Discard, reader)

	containerName := fmt.Sprintf("bifrost_%s", namesgenerator.GetRandomName(1))

	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:      image,
			Entrypoint: entrypoint,
			Cmd:        cmd,
		},
		&container.HostConfig{
			Mounts: mounts,
		},
		nil, containerName,
	)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("created container %s:%s", containerName, resp.ID[0:12])

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	logrus.Debugf("started container %s:%s", containerName, resp.ID[0:12])

	return ContainerFromId(resp.ID), nil

}

// ID returns the current id
func (c *Container) ID() string {
	return c.id
}

// RunCommand executes the given command with arguments inside the container
func (c *Container) RunCommand(command []string) (output string, err error)  {
	return RunCommand(c.id, command)
}

// RunShellCommand executes the given command string within a /bin/bash shell inside the container
func (c *Container) RunShellCommand(command string) (output string, err error)  {
	return RunShellCommand(c.id, command)
}

// Remove removes the container.
// force allows a running container to be forcefully removed.
func (c *Container) Remove(force bool) (err error)  {
	return RemoveContainer(c.id, force)
}

// CopyTo copies the given contents to the container
func (c *Container) CopyTo(path string, contents io.Reader) error {
	return CopyToContainer(c.id, path, contents)
}