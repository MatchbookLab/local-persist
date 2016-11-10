package main

import (
    "fmt"
    "os/user"

    "github.com/docker/go-plugins-helpers/volume"
)

func main() {
    driver := newLocalPersistDriver()

    u, _ := user.Current()

    handler := volume.NewHandler(driver)
    fmt.Println(handler.ServeUnix(u.Name, driver.name))
}
