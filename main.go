package main

import (
	"fmt"
	"os"
	"os/user"
	"strconv"

  "github.com/Carbonique/local-persist/driver"
	"github.com/docker/go-plugins-helpers/volume"
)

const (
	stateDir = "/local-persist/state"
	dataDir  = "/local-persist/data"
)

func main() {

	d, err := driver.NewLocalPersistDriver(stateDir, dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	u, _ := user.Lookup("root")
	uid, _ := strconv.Atoi(u.Uid)

	handler := volume.NewHandler(d)
	fmt.Println(handler.ServeUnix(d.Name, uid))
}
