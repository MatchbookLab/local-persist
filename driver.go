package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
)

const (
	stateDir  = "/var/lib/docker/plugin-data/"
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

func newLocalPersistDriver() localPersistDriver {
	log.Println("Starting new Driver")

	driver := localPersistDriver{
		volumes: map[string]string{},
		mutex:   &sync.Mutex{},
		debug:   true,
		name:    "local-persist",
	}

	os.Mkdir(stateDir, 0700)

	driver.volumes, _ = driver.findExistingVolumesFromStateFile()
	log.Printf("found %d volumes on startup \n", len(driver.volumes))

	return driver
}

func (driver localPersistDriver) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	log.Println("New get request")

	if !driver.exists(req.Name) {
		return &volume.GetResponse{}, fmt.Errorf("no volume found with name %s", req.Name)
	}

	return &volume.GetResponse{
		Volume: driver.volume(req.Name),
	}, nil

}

func (driver localPersistDriver) List() (*volume.ListResponse, error) {
	log.Println("List called")

	var volumes []*volume.Volume

	for name := range driver.volumes {
		volumes = append(volumes, driver.volume(name))
	}

	return &volume.ListResponse{
		Volumes: volumes,
	}, nil
}

func (driver localPersistDriver) Create(req *volume.CreateRequest) error {
	log.Println("Create called")

	mountpoint := req.Options["mountpoint"]
	if mountpoint == "" {
		return fmt.Errorf("the `mountpoint` option is required")
	}

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	if driver.exists(req.Name) {
		return fmt.Errorf("the volume %s already exists", req.Name)
	}

	err := os.MkdirAll(mountpoint, 0755)
	log.Printf("Ensuring directory %s exists on host...\n", mountpoint)

	if err != nil {
		log.Printf("Could not create directory %s\n", mountpoint)
		return errors.New("error on create. Could not create directory")
	}

	driver.volumes[req.Name] = mountpoint
	e := driver.saveState(driver.volumes)
	if e != nil {
		fmt.Println(e.Error())
	}

	log.Printf("Created volume %s with mountpoint %s \n", req.Name, mountpoint)

	return err
}

func (driver localPersistDriver) Remove(req *volume.RemoveRequest) error {
	log.Println("Remove called")
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	delete(driver.volumes, req.Name)

	err := driver.saveState(driver.volumes)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("Removed %s\n", req.Name)

	return err
}

func (driver localPersistDriver) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	log.Println("Mount called")

	return &volume.MountResponse{Mountpoint: driver.volumes[req.Name]}, nil
}

func (driver localPersistDriver) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	log.Println("Path called")

	fmt.Printf("Returned path %s\n", driver.volumes[req.Name])

	return &volume.PathResponse{Mountpoint: driver.volumes[req.Name]}, nil
}

func (driver localPersistDriver) Unmount(req *volume.UnmountRequest) error {
	log.Println("Unmount Called... ")

	log.Printf("Unmounted %s\n", req.Name)

	return nil
}

func (driver localPersistDriver) Capabilities() *volume.CapabilitiesResponse {
	log.Println("Capabilities Called... ")

	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{Scope: "local"},
	}
}

func (driver localPersistDriver) exists(name string) bool {
	return driver.volumes[name] != ""
}

func (driver localPersistDriver) volume(name string) *volume.Volume {
	return &volume.Volume{
		Name: name,

		Mountpoint: driver.volumes[name],
	}
}

func (driver localPersistDriver) findExistingVolumesFromStateFile() (map[string]string, error) {
	path := path.Join(stateDir, stateFile)
	fileData, err := ioutil.ReadFile(path)
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

func (driver localPersistDriver) saveState(volumes map[string]string) error {
	data := saveData{
		State: volumes,
	}

	fileData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	path := path.Join(stateDir, stateFile)
	return ioutil.WriteFile(path, fileData, 0600)
}
