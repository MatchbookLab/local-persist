package main

import (
    "fmt"
    "sync"
    "os"

    "github.com/docker/go-plugins-helpers/volume"
    "github.com/docker/engine-api/client"
    "github.com/docker/engine-api/types"
)

type localPersistDriver struct {
    volumes    map[string]string
    mutex      *sync.Mutex
    debug      bool
    name       string
}

func newLocalPersistDriver() localPersistDriver {
    driver := localPersistDriver{
        volumes : map[string]string{},
		mutex   : &sync.Mutex{},
        debug   : true,
        name    : "local-persist",
    }

    // set up the ability to make API calls to the daemon
    defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
    cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.21", nil, defaultHeaders)
    if err != nil {
        panic(err)
    }

    // grab ALL containers...
    options := types.ContainerListOptions{All: true}
    containers, err := cli.ContainerList(options)

    // ...and check to see if any of them belong to this driver and recreate their references
    for _, container := range containers {
        info, err := cli.ContainerInspect(container.ID)
        if err != nil {
            panic(err)
        }

        for _, mount := range info.Mounts {
            if mount.Driver == driver.name {
                // @TODO there could be multiple volumes (mounts) with this { name: source } combo, and while that's okay
                // what if they is the same name with a different source? could that happen? if it could,
                // it'd be bad, so maybe we want to panic here?
                driver.volumes[mount.Name] = mount.Source
            }
        }
    }

    fmt.Printf("Found %d volumes on startup\n", len(driver.volumes))

    return driver
}

func (driver localPersistDriver) Get(req volume.Request) volume.Response {
    fmt.Print("Get Called... ")

    if driver.exists(req.Name) {
        fmt.Printf("[%s] was found\n", req.Name)
        return volume.Response{
            Volume: driver.volume(req.Name),
        }
    }

    fmt.Printf("[%s] not found\n", req.Name)
    return volume.Response{
        Err: fmt.Sprintf("No volume found with the name [%s]", req.Name),
    }

}

func (driver localPersistDriver) List(req volume.Request) volume.Response {
    fmt.Print("List Called... ")

    var volumes []*volume.Volume
    for name, _ := range driver.volumes {
        volumes = append(volumes, driver.volume(name))
    }

    fmt.Printf("Found %d volumes\n", len(volumes))

    return volume.Response{
        Volumes: volumes,
    }
}

func (driver localPersistDriver) Create(req volume.Request) volume.Response {
    fmt.Println("Create Called...")
    mountpoint := req.Options["mountpoint"]
    if mountpoint == "" {
        return volume.Response{ Err: "The `mountpoint` option is required" }
    }

    driver.mutex.Lock()
    defer driver.mutex.Unlock()

    err := os.MkdirAll(mountpoint, 0755)

    if (err != nil) {
        return volume.Response{ Err: err.Error() }
    }

    driver.volumes[req.Name] = mountpoint

    fmt.Printf("Creating volume [%s] mounted at [%s]\n", req.Name, mountpoint)

    return volume.Response{}
}

func (driver localPersistDriver) Remove(req volume.Request) volume.Response {
    fmt.Println("Remove Called...")
    driver.mutex.Lock()
    defer driver.mutex.Unlock()

    delete(driver.volumes, req.Name)

    return volume.Response{}
}

func (driver localPersistDriver) Mount(req volume.Request) volume.Response {
    fmt.Println("Mount Called...")

    return driver.Path(req)
}

func (driver localPersistDriver) Path(req volume.Request) volume.Response {
    fmt.Println("Path Called...")

    return volume.Response{ Mountpoint:  driver.volumes[req.Name] }
}

func (driver localPersistDriver) Unmount(req volume.Request) volume.Response {
    fmt.Println("Unmount Called...")

    return driver.Path(req)
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
