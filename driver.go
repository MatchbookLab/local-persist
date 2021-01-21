package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/fatih/color"
	"golang.org/x/net/context"
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

func newLocalPersistDriver() localPersistDriver {
	fmt.Printf(white("%-18s", "Starting... "))

	driver := localPersistDriver{
		volumes: map[string]string{},
		mutex:   &sync.Mutex{},
		debug:   true,
		name:    "local-persist",
	}

	os.Mkdir(stateDir, 0700)

	_, driver.volumes = driver.findExistingVolumesFromStateFile()
	fmt.Printf("Found %s volumes on startup\n", yellow(strconv.Itoa(len(driver.volumes))))

	return driver
}

func (driver localPersistDriver) Get(req volume.Request) volume.Response {
	fmt.Print(white("%-18s", "Get Called... "))

	if driver.exists(req.Name) {
		fmt.Printf("Found %s\n", cyan(req.Name))
		return volume.Response{
			Volume: driver.volume(req.Name),
		}
	}

	fmt.Printf("Couldn't find %s\n", cyan(req.Name))
	return volume.Response{
		Err: fmt.Sprintf("No volume found with the name %s", cyan(req.Name)),
	}
}

func (driver localPersistDriver) List(req volume.Request) volume.Response {
	fmt.Print(white("%-18s", "List Called... "))

	var volumes []*volume.Volume
	for name, _ := range driver.volumes {
		volumes = append(volumes, driver.volume(name))
	}

	fmt.Printf("Found %s volumes\n", yellow(strconv.Itoa(len(volumes))))

	return volume.Response{
		Volumes: volumes,
	}
}

func (driver localPersistDriver) Create(req volume.Request) volume.Response {
	fmt.Print(white("%-18s", "Create Called... "))

	mountpoint := req.Options["mountpoint"]
	if mountpoint == "" {
		fmt.Printf("No %s option provided\n", blue("mountpoint"))
		return volume.Response{Err: fmt.Sprintf("The `mountpoint` option is required")}
	}

	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	if driver.exists(req.Name) {
		return volume.Response{Err: fmt.Sprintf("The volume %s already exists", req.Name)}
	}

	err := os.MkdirAll(mountpoint, 0755)
	fmt.Printf("Ensuring directory %s exists on host...\n", magenta(mountpoint))

	if err != nil {
		fmt.Printf("%17s Could not create directory %s\n", " ", magenta(mountpoint))
		return volume.Response{Err: err.Error()}
	}

	driver.volumes[req.Name] = mountpoint
	e := driver.saveState(driver.volumes)
	if e != nil {
		fmt.Println(e.Error())
	}

	fmt.Printf("%17s Created volume %s with mountpoint %s\n", " ", cyan(req.Name), magenta(mountpoint))

	return volume.Response{}
}

func (driver localPersistDriver) Remove(req volume.Request) volume.Response {
	fmt.Print(white("%-18s", "Remove Called... "))
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	delete(driver.volumes, req.Name)

	err := driver.saveState(driver.volumes)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("Removed %s\n", cyan(req.Name))

	return volume.Response{}
}

func (driver localPersistDriver) Mount(req volume.MountRequest) volume.Response {
	fmt.Print(white("%-18s", "Mount Called... "))

	fmt.Printf("Mounted %s\n", cyan(req.Name))

	return driver.Path(volume.Request{Name: req.Name})
}

func (driver localPersistDriver) Path(req volume.Request) volume.Response {
	fmt.Print(white("%-18s", "Path Called... "))

	fmt.Printf("Returned path %s\n", magenta(driver.volumes[req.Name]))

	return volume.Response{Mountpoint: driver.volumes[req.Name]}
}

func (driver localPersistDriver) Unmount(req volume.UnmountRequest) volume.Response {
	fmt.Print(white("%-18s", "Unmount Called... "))

	fmt.Printf("Unmounted %s\n", cyan(req.Name))

	return driver.Path(volume.Request{Name: req.Name})
}

func (driver localPersistDriver) Capabilities(req volume.Request) volume.Response {
	fmt.Print(white("%-18s", "Capabilities Called... "))

	return volume.Response{
		Capabilities: volume.Capability{Scope: "local"},
	}
}

func (driver localPersistDriver) exists(name string) bool {
	return driver.volumes[name] != ""
}

func (driver localPersistDriver) volume(name string) *volume.Volume {
	return &volume.Volume{
		Name:       name,
		Mountpoint: driver.volumes[name],
	}
}

func (driver localPersistDriver) findExistingVolumesFromDockerDaemon() (error, map[string]string) {
	// set up the ability to make API calls to the daemon
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	// need at least Docker 1.9 (API v1.21) for named Volume support
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.21", nil, defaultHeaders)
	if err != nil {
		return err, map[string]string{}
	}

	// grab ALL containers...
	options := types.ContainerListOptions{All: true}
	containers, err := cli.ContainerList(context.Background(), options)

	// ...and check to see if any of them belong to this driver and recreate their references
	var volumes = map[string]string{}
	for _, container := range containers {
		info, err := cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			// something really weird happened here... PANIC
			panic(err)
		}

		for _, mount := range info.Mounts {
			if mount.Driver == driver.name {
				// @TODO there could be multiple volumes (mounts) with this { name: source } combo, and while that's okay
				// what if they is the same name with a different source? could that happen? if it could,
				// it'd be bad, so maybe we want to panic here?
				volumes[mount.Name] = mount.Source
			}
		}
	}

	if err != nil || len(volumes) == 0 {
		fmt.Print("Attempting to load from file state...   ")

		return driver.findExistingVolumesFromStateFile()
	}

	return nil, volumes
}

func (driver localPersistDriver) findExistingVolumesFromStateFile() (error, map[string]string) {
	filePath := path.Join(stateDir, stateFile)
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err, map[string]string{}
	}

	var data saveData
	e := json.Unmarshal(fileData, &data)
	if e != nil {
		return e, map[string]string{}
	}

	return nil, data.State
}

func (driver localPersistDriver) saveState(volumes map[string]string) error {
	data := saveData{
		State: volumes,
	}

	fileData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	filePath := path.Join(stateDir, stateFile)
	return ioutil.WriteFile(filePath, fileData, 0600)
}
