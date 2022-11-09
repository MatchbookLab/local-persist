package main

import (
	"fmt"
	"os/user"
	"strconv"

	"github.com/docker/go-plugins-helpers/volume"
)

func main() {
	driver := newLocalPersistDriver()

	u, _ := user.Lookup("root")
	uid, _ := strconv.Atoi(u.Uid)

	handler := volume.NewHandler(driver)
	fmt.Println(handler.ServeUnix(driver.name, uid))
}
