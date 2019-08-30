package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

type dockerClient interface {
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error)
}

var (
	logReaderOptions = types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
		Details:    false,
	}
	containerListOptions = types.ContainerListOptions{
		All: true,
	}
	eventSubscribeOptions = types.EventsOptions{
		Since:   "", // TODO: Needs to be stored to only get the events since the last time checked
		Filters: filters.NewArgs(filters.KeyValuePair{Key: eventsToSubscribe, Value: "true"}),
	}
)

func newClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, errors.Wrap(err, "Failed getting Docker client")
	}
	return cli, nil
}

func runningContainers(client dockerClient) ([]types.Container, error) {
	containers, err := client.ContainerList(context.Background(), containerListOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Failed getting list of Docker containers")
	}
	return containers, nil
}

func containerLogs(client dockerClient, containerID string) (io.ReadCloser, error) {
	reader, err := client.ContainerLogs(context.Background(), containerID, logReaderOptions)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed getting logs for container_id %s", containerID)
	}
	return reader, nil
}

func listenToEvents(client dockerClient) (<-chan events.Message, <-chan error) {
	return client.Events(context.Background(), eventSubscribeOptions)
}
