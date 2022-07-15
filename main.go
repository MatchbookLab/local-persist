package main

import (
	"os/user"
	"strconv"

	"github.com/docker/go-plugins-helpers/volume"
)

func main() {
	d := newLocalPersistDriver()

	h := volume.NewHandler(d)
	u, _ := user.Lookup("root")
	gid, _ := strconv.Atoi(u.Gid)
	h.ServeUnix(d.name, gid)
}
