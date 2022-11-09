package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/sirupsen/logrus"
)

const (
	stateDir  = "/state"
	stateFile = "local-persist.json"
)

type localPersistDriver struct {
	volumes map[string]string
	mutex   *sync.Mutex
	debug   bool
	name    string
}

type saveData struct {
	State map[string]string `json:"state"`
}

func newLocalPersistDriver() *localPersistDriver {
	log.Info("Starting")

	driver := localPersistDriver{
		volumes: map[string]string{},
		mutex:   &sync.Mutex{},
		debug:   true,
		name:    "local-persist",
	}

	os.Mkdir(stateDir, 0700)

	driver.volumes, _ = driver.findExistingVolumesFromStateFile()
	log.Infof("Found %d volumes on startup\n", len(driver.volumes))
	return &driver
}

func (driver *localPersistDriver) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	log.Debug("Get called")

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	if !driver.exists(req.Name) {
		log.Errorf("Couldn't find %s\n", req.Name)

		return &volume.GetResponse{}, fmt.Errorf("no volume found with the name %s", req.Name)
	}

	log.Infof("Found %s\n", req.Name)

	return &volume.GetResponse{Volume: driver.volume(req.Name)}, nil
}

func (driver *localPersistDriver) List() (*volume.ListResponse, error) {
	log.Debug("List called")

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	var volumes []*volume.Volume
	for name := range driver.volumes {
		volumes = append(volumes, driver.volume(name))
	}

	log.Infof("Found %d volumes on startup\n", len(driver.volumes))

	return &volume.ListResponse{Volumes: volumes}, nil
}

func (driver *localPersistDriver) Create(req *volume.CreateRequest) error {
	log.Debug("Create called")

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	mountpoint := req.Options["mountpoint"]
	if mountpoint == "" {
		log.Infof("No %s option provided\n", "mountpoint")
		return fmt.Errorf("the `mountpoint` option is required")
	}

	if driver.exists(req.Name) {
		return fmt.Errorf("the volume %s already exists", req.Name)
	}

	err := os.MkdirAll(mountpoint, 0755)
	log.Infof("Ensuring directory %s exists on host...\n", mountpoint)

	if err != nil {
		return fmt.Errorf("%17s could not create directory %s", " ", mountpoint)
	}

	driver.volumes[req.Name] = mountpoint
	e := driver.saveState(driver.volumes)
	if e != nil {
		return fmt.Errorf("error %s", e)
	}

	log.Infof("Created volume %s with mountpoint %s", req.Name, mountpoint)

	return nil
}

func (driver *localPersistDriver) Remove(req *volume.RemoveRequest) error {
	log.Debug("Remove called")

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	delete(driver.volumes, req.Name)

	err := driver.saveState(driver.volumes)
	if err != nil {
		return fmt.Errorf("error %s", err)
	}

	log.Infof("Removed %s", req.Name)

	return nil
}

func (driver *localPersistDriver) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	log.Debug("Mount called")

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	p, ok := driver.volumes[req.Name]

	if !ok {
		return &volume.MountResponse{}, fmt.Errorf("volume %s not found", req.Name)
	} // Now check if the path still exists on the host
	f, err := os.Stat(p)

	// If the path does not exist
	if errors.Is(err, fs.ErrNotExist) {
		return &volume.MountResponse{}, fmt.Errorf("Path %s for volume %s not found", p, req.Name)
	}

	// If the path is a file
	if f != nil && !f.IsDir() {
		return &volume.MountResponse{}, fmt.Errorf("Path %s for volume %s is a file, not a directory", p, req.Name)
	}

	log.Infof("Mounted %s", req.Name)

	return &volume.MountResponse{Mountpoint: p}, nil
}

func (driver *localPersistDriver) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	log.Debug("Mount called")

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	v, ok := driver.volumes[req.Name]
	if !ok {
		return &volume.PathResponse{}, fmt.Errorf("volume %s not found", req.Name)
	}
	log.Infof("Returned path %s", v)

	return &volume.PathResponse{Mountpoint: driver.volumes[req.Name]}, nil
}

func (driver *localPersistDriver) Unmount(req *volume.UnmountRequest) error {
	log.Debug("Unmount called")

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	_, ok := driver.volumes[req.Name]
	if !ok {
		return fmt.Errorf("volume %s not found", req.Name)
	}

	log.Infof("Unmounted %s", req.Name)

	return nil
}

func (driver *localPersistDriver) Capabilities() *volume.CapabilitiesResponse {
	log.Debug("Capabilities called")

	return &volume.CapabilitiesResponse{Capabilities: volume.Capability{Scope: "local"}}
}

func (driver *localPersistDriver) exists(name string) bool {
	return driver.volumes[name] != ""
}

func (driver *localPersistDriver) volume(name string) *volume.Volume {
	return &volume.Volume{
		Name:       name,
		Mountpoint: driver.volumes[name],
	}
}

func (driver *localPersistDriver) findExistingVolumesFromStateFile() (map[string]string, error) {
	path := path.Join(stateDir, stateFile)
	fileData, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}, err
	}

	var data saveData
	e := json.Unmarshal(fileData, &data)
	if e != nil {
		return map[string]string{}, e
	}

	return data.State, nil
}

func (driver *localPersistDriver) saveState(volumes map[string]string) error {
	data := saveData{
		State: volumes,
	}

	fileData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	path := path.Join(stateDir, stateFile)
	return os.WriteFile(path, fileData, 0600)
}
