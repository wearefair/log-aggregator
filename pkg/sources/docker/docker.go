package docker

import (
	"log"
	"sync"

	"github.com/docker/docker/api/types/events"
	"github.com/pkg/errors"
	"github.com/wearefair/log-aggregator/pkg/types"
)

// Docker reads Docker log files in JSON format
type Docker struct {
	client        dockerClient
	out           chan<- *types.Record
	containers    map[string]*container
	containerLock *sync.Mutex
}

// NewDocker returns a new Docker log reader
func NewDocker() (*Docker, error) {
	cli, err := newClient()
	if err != nil {
		return nil, errors.Wrap(err, "Failed getting Docker client")
	}

	return &Docker{
		client:     cli,
		containers: make(map[string]*container),
	}, nil
}

// Start starts the Docker reader
func (docker *Docker) Start(out chan<- *types.Record) {
	docker.out = out
	docker.initContainers()
	docker.watchEvents()
}

// Stop stops the Docker reader
func (docker *Docker) Stop() {
	containers := make(map[string]struct{})
	docker.containerLock.Lock()
	for i := range docker.containers {
		containers[i] = struct{}{}
	}
	docker.containerLock.Unlock()

	for container := range containers {
		if err := docker.removeContainer(container); err != nil {
			log.Printf("Failed removing container with ID %s. Error %s", container, err)
		}
	}
}

func (docker *Docker) addContainer(containerID string) error {
	rdr, err := containerLogs(docker.client, containerID)
	if err != nil {
		return errors.Wrap(err, "Failed getting logs")
	}

	container := &container{
		id:  containerID,
		rdr: rdr,
		out: docker.out,
	}
	go container.read()

	docker.containerLock.Lock()
	docker.containers[containerID] = container
	docker.containerLock.Unlock()

	return nil
}

func (docker *Docker) removeContainer(containerID string) error {
	docker.containerLock.Lock()
	defer docker.containerLock.Unlock()

	var container *container
	var ok bool
	if container, ok = docker.containers[containerID]; !ok {
		return errors.New("Container not found")
	}

	if err := container.stop(); err != nil {
		return errors.Wrapf(err, "Failed stopping to reader the container %s logs", containerID)
	}
	delete(docker.containers, containerID)
	return nil
}

func (docker *Docker) initContainers() error {
	containers, err := runningContainers(docker.client)
	if err != nil {
		return errors.Wrap(err, "Failed refreshing container list")
	}

	for _, container := range containers {
		docker.addContainer(container.ID)
	}
	return nil
}

func (docker *Docker) watchEvents() {
	message, errChan := listenToEvents(docker.client)
	go func() {
		err := <-errChan
		if err != nil {
			log.Printf("Error while listening for Docker events. Error %s", err)
		}
	}()
	for {
		docker.processEvent(<-message)
	}
}

func (docker *Docker) processEvent(msg events.Message) {
	if containerID, ok := isContainerCreation(msg); ok {
		docker.addContainer(containerID)
	}
	if containerID, ok := isContainerDestruction(msg); ok {
		docker.removeContainer(containerID)
	}
}
