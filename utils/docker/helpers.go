package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"io"
	"io/ioutil"
)

// RemoveContainer removes the container with containerId.
// force allows a running container to be forcefully removed.
func RemoveContainer(containerId string, force bool) error {
	docker := GetClient()
	ctx := context.Background()
	return docker.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{
		Force: force,
	})
}

// CopyToContainer copies the given contents to the container with containerId
func CopyToContainer(containerId string, path string, contents io.Reader) error {
	docker := GetClient()
	ctx := context.Background()
	return docker.CopyToContainer(ctx, containerId, path, contents, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	})
}

// RunCommand executes the given command with arguments inside the container with containerId
func RunCommand(containerId string, command []string) (string, error) {
	docker := GetClient()
	ctx := context.Background()

	execId, err := docker.ContainerExecCreate(ctx, containerId, types.ExecConfig{
		Cmd:          command,
		Detach:       false,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return "", err
	}
	res, err := docker.ContainerExecAttach(ctx, execId.ID, types.ExecConfig{})
	if err != nil {
		return "", err
	}
	outputBytes, err := ioutil.ReadAll(res.Reader)
	if err != nil {
		return "", err
	}

	info, err := docker.ContainerExecInspect(ctx, execId.ID)
	if err != nil {
		return "", err
	}

	if info.ExitCode != 0 {
		return string(outputBytes), fmt.Errorf("`%s` exited with code %d", command, info.ExitCode)
	}

	return string(outputBytes), nil
}

// RunShellCommand executes the given command string within a /bin/bash shell inside the container with containerId
func RunShellCommand(containerId string, command string) (string, error) {
	return RunCommand(containerId, []string{"/bin/bash", "-c", command})
}
