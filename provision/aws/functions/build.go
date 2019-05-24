package functions

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/niranjan94/bifrost/config"
	"github.com/niranjan94/bifrost/utils"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
)

var containerIds []string

func getContainerFor(runtime string) (string, error) {
	ctx := context.Background()
	cli := utils.GetDockerClient()

	image := fmt.Sprintf("lambci/lambda:%s", runtime)
	canonicalPath := fmt.Sprintf("docker.io/%s", image)

	logrus.Debug("looking up image ", canonicalPath)

	reader, err := cli.ImagePull(ctx, canonicalPath, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}

	logrus.Debug("pulling image ", canonicalPath)
	io.Copy(ioutil.Discard, reader)

	containerName := fmt.Sprintf("bifrost_%s", namesgenerator.GetRandomName(1))

	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image: image,
			Entrypoint: []string{"tail"},
			Cmd:   []string{"-f", "/dev/null"},
			Volumes: map[string]struct{}{},
		}, nil, nil, containerName,
	)
	if err != nil {
		return "", err
	}

	logrus.Debugf("created container %s:%s", containerName, resp.ID[0:12])

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	logrus.Debugf("started container %s:%s", containerName, resp.ID[0:12])

	return resp.ID, nil
}

func cleanupContainers()  {
	docker := utils.GetDockerClient()
	ctx := context.Background()
	logrus.Debug("cleaning up containers")
	for _, id := range containerIds {
		if err := docker.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			logrus.Error(err)
		}
	}
}

func Build() {
	functionsMap := config.GetStringMapSub("serverless.functions", true)
	defer cleanupContainers()
	for k, v := range functionsMap {
		logrus.Info(k)
		containerId, err := getContainerFor(v.GetString("runtime"))
		if err != nil {
			logrus.Error(err)
		}
		containerIds = append(containerIds, containerId)
	}
}
