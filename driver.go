package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/fatih/color"
)

var (
	// red = color.New(color.FgRed).SprintfFunc()
	// green = color.New(color.FgGreen).SprintfFunc()
	yellow  = color.New(color.FgYellow).SprintfFunc()
	cyan    = color.New(color.FgCyan).SprintfFunc()
	blue    = color.New(color.FgBlue).SprintfFunc()
	magenta = color.New(color.FgMagenta).SprintfFunc()
	white   = color.New(color.FgWhite).SprintfFunc()
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

func newLocalPersistDriver() *localPersistDriver {
	fmt.Print("\n", white("%-18s", "Starting... "))

	driver := localPersistDriver{
		volumes: map[string]string{},
		mutex:   &sync.Mutex{},
		debug:   true,
		name:    "local-persist",
	}

	os.Mkdir(stateDir, 0700)

	driver.volumes, _ = driver.findExistingVolumesFromStateFile()
	fmt.Printf("Found %s volumes on startup\n", yellow(strconv.Itoa(len(driver.volumes))))

	return &driver
}

func (driver *localPersistDriver) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	fmt.Print("\n", white("%-18s", "Get Called... "))

	if !driver.exists(req.Name) {
		fmt.Printf("Couldn't find %s\n", cyan(req.Name))
		return nil, fmt.Errorf("no volume found with the name %s", cyan(req.Name))
	}

	fmt.Printf("Found %s\n", cyan(req.Name))

	return &volume.GetResponse{Volume: driver.volume(req.Name)}, nil
}

func (driver *localPersistDriver) List() (*volume.ListResponse, error) {
	fmt.Print("\n", white("%-18s", "List Called... "))

	var volumes []*volume.Volume
	for name := range driver.volumes {
		volumes = append(volumes, driver.volume(name))
	}

	fmt.Printf("Found %s volumes\n", yellow(strconv.Itoa(len(volumes))))

	return &volume.ListResponse{Volumes: volumes}, nil
}

func (driver *localPersistDriver) Create(req *volume.CreateRequest) error {
	fmt.Print("\n", white("%-18s", "Create Called... "))

	mountpoint := req.Options["mountpoint"]
	if mountpoint == "" {
		fmt.Printf("No %s option provided\n", blue("mountpoint"))
		return fmt.Errorf("the `mountpoint` option is required")
	}

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	if driver.exists(req.Name) {
		return fmt.Errorf("the volume %s already exists", req.Name)
	}

	err := os.MkdirAll(mountpoint, 0755)
	fmt.Printf("Ensuring directory %s exists on host...\n", magenta(mountpoint))

	if err != nil {
		return fmt.Errorf("%17s could not create directory %s", " ", magenta(mountpoint))
	}

	driver.volumes[req.Name] = mountpoint
	e := driver.saveState(driver.volumes)
	if e != nil {
		return fmt.Errorf("error %s", e)
	}

	fmt.Printf("%17s Created volume %s with mountpoint %s", " ", cyan(req.Name), magenta(mountpoint))

	return nil
}

func (driver *localPersistDriver) Remove(req *volume.RemoveRequest) error {
	fmt.Print("\n", white("%-18s", "Remove Called... "))
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	delete(driver.volumes, req.Name)

	err := driver.saveState(driver.volumes)
	if err != nil {
		return fmt.Errorf("error %s", err)
	}

	fmt.Printf("Removed %s\n", cyan(req.Name))

	return nil
}

func (driver *localPersistDriver) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	// TODO: Improve error handling. What if volume exists, but mountpoint/path has been deleted?
	fmt.Print("\n", white("%-18s", "Mount Called... "))

	_, ok := driver.volumes[req.Name]
	if !ok {
		return &volume.MountResponse{}, fmt.Errorf("volume %s not found", req.Name)
	}

	fmt.Printf("Mounted %s\n", cyan(req.Name))

	return &volume.MountResponse{Mountpoint: driver.volumes[req.Name]}, nil
}

func (driver *localPersistDriver) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	// TODO: Improve error handling, what if path no longer exists, what if volume does not exist?

	fmt.Print("\n", white("%-18s", "Path Called... "))

	fmt.Printf("Returned path %s\n", magenta(driver.volumes[req.Name]))

	return &volume.PathResponse{Mountpoint: driver.volumes[req.Name]}, nil
}

func (driver *localPersistDriver) Unmount(req *volume.UnmountRequest) error {
	// TODO: Improve error handling. What if volume is not found?
	// And: Is this function even doing anything? What should it be doing?
	fmt.Print("\n", white("%-18s", "Unmount Called... "))

	fmt.Printf("Unmounted %s\n", cyan(req.Name))

	return nil
}

func (driver *localPersistDriver) Capabilities() *volume.CapabilitiesResponse {
	fmt.Print("\n", white("%-18s", "Capabilities Called... "))

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

//func (driver localPersistDriver) findExistingVolumesFromDockerDaemon() (map[string]string, error) {
//	// set up the ability to make API calls to the daemon
//	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
//	// need at least Docker 1.9 (API v1.21) for named Volume support
//	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.21", nil, defaultHeaders)
//	if err != nil {
//		return map[string]string{}, err
//	}
//
//	// grab ALL containers...
//	options := types.ContainerListOptions{All: true}
//	containers, err := cli.ContainerList(context.Background(), options)
//
//	// ...and check to see if any of them belong to this driver and recreate their references
//	var volumes = map[string]string{}
//	for _, container := range containers {
//		info, err := cli.ContainerInspect(context.Background(), container.ID)
//		if err != nil {
//			// something really weird happened here... PANIC
//			panic(err)
//		}
//
//		for _, mount := range info.Mounts {
//			if mount.Driver == driver.name {
//				// @TODO there could be multiple volumes (mounts) with this { name: source } combo, and while that's okay
//				// what if they is the same name with a different source? could that happen? if it could,
//				// it'd be bad, so maybe we want to panic here?
//				volumes[mount.Name] = mount.Source
//			}
//		}
//	}
//
//	if err != nil || len(volumes) == 0 {
//		fmt.Print("\n","Attempting to load from file state...   ")
//
//		return driver.findExistingVolumesFromStateFile()
//	}
//
//	return volumes, nil
//}

func (driver *localPersistDriver) findExistingVolumesFromStateFile() (map[string]string, error) {
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

func (driver *localPersistDriver) saveState(volumes map[string]string) error {
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
