package utils

import (
	"github.com/docker/docker/client"
	"sync"
)

var (
	once sync.Once
	dockerClient *client.Client
)

func GetDockerClient() *client.Client {
	once.Do(func() {
		cli, err := client.NewEnvClient()
		if err != nil {
			panic(err)
		}
		dockerClient = cli
	})
	return dockerClient
}