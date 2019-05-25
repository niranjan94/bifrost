package docker

import (
	"github.com/docker/docker/client"
	"sync"
)

var (
	once sync.Once
	dockerClient *client.Client
)

// GetClient returns a new docker client created from the current environment
// the client is created if not already
func GetClient() *client.Client {
	once.Do(func() {
		cli, err := client.NewEnvClient()
		if err != nil {
			panic(err)
		}
		dockerClient = cli
	})
	return dockerClient
}