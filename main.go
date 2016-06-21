package main

import (
    "fmt"
    "os"
    "github.com/urfave/cli"
    "github.com/docker/go-plugins-helpers/volume"
)

func main() {
    var name string
    var baseDir string
    var stateDir string

	app := cli.NewApp()
    app.Name = "docker-local-persist"
    app.Version = "1.2.2"
    app.Usage = "Local Persist Volume Plugin for Docker"
    app.Flags = []cli.Flag {
        cli.StringFlag{
            Name: "name",
            Value: "local-persist",
            Usage: "docker plugin name",
            Destination: &name,
        },
        cli.StringFlag{
            Name: "baseDir",
            Value: "",
            Usage: "Mounted volume base directory",
            Destination: &baseDir,
        },
        cli.StringFlag{
            Name: "stateDir",
            Value: "/var/lib/docker/plugin-data/",
            Usage: "state directory",
            Destination: &stateDir,
        },
    }

    app.Action = func(c *cli.Context) error {
        driver := newLocalPersistDriver(name, baseDir, stateDir)
        handler := volume.NewHandler(driver)
        fmt.Println(handler.ServeUnix("root", driver.name))
        return nil
    }
	app.Run(os.Args)
}
