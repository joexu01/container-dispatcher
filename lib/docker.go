package lib

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type DockerClient struct {
	Ctx    context.Context
	Client *client.Client
}

func NewDockerClient(listKey string) (*DockerClient, error) {
	//log.Debug("What! %s\n\n", ConfDockerMap.List[listKey].DockerEndpoint)
	cli, err := client.NewClientWithOpts(client.WithHost(ConfDockerMap.List[listKey].DockerEndpoint), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerClient{
		Ctx:    context.Background(),
		Client: cli,
	}, nil
}

func (d *DockerClient) ListAllImages() ([]types.ImageSummary, error) {
	images, err := d.Client.ImageList(d.Ctx, types.ImageListOptions{
		All:     true,
		Filters: filters.Args{},
	})
	if err != nil {
		return nil, err
	}
	return images, nil
}

func (d *DockerClient) ListContainers(all bool) ([]types.Container, error) {
	containers, err := d.Client.ContainerList(d.Ctx, types.ContainerListOptions{
		Quiet:   false,
		Size:    false,
		All:     all,
		Latest:  true,
		Since:   "",
		Before:  "",
		Limit:   0,
		Filters: filters.Args{},
	})
	if err != nil {
		return nil, err
	}

	return containers, nil
}
