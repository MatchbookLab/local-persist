package main

import (
	"os"
	"testing"

	"github.com/docker/go-plugins-helpers/volume"
)

var (
	defaultTestName       = "test-volume"
	defaultTestMountpoint = "./tmp/data/local-persist-test"
)

func TestCreate(t *testing.T) {
	driver := newLocalPersistDriver()

	defaultCreateHelper(driver, t)

	// test that a directory is created
	_, err := os.Stat(defaultTestMountpoint)
	if os.IsNotExist(err) {
		t.Error("Mountpoint directory was not created:", err.Error())
	}

	// test that volumes has one
	if len(driver.volumes) != 1 {
		t.Error("Driver should have exactly 1 volume")
	}

	defaultCleanupHelper(driver, t)

	req := &volume.CreateRequest{Name: defaultTestName}
	// test that options are required
	err = driver.Create(req)

	if err == nil {
		t.Error("No error was returned, although we did not pass mountpoint")
	}
}

func TestGet(t *testing.T) {
	driver := newLocalPersistDriver()

	defaultCreateHelper(driver, t)

	req := &volume.GetRequest{Name: defaultTestName}

	res, err := driver.Get(req)
	if err != nil {
		t.Error("Should have found a volume!")
	}

	if res.Volume.Name != defaultTestName {
		t.Error("Incorrect volume name was returned")
	}

	defaultCleanupHelper(driver, t)
}

func TestList(t *testing.T) {
	driver := newLocalPersistDriver()

	name := defaultTestName + "2"
	mountpoint := defaultTestMountpoint + "2"

	defaultCreateHelper(driver, t)

	res, err := driver.List()

	if err != nil {
		t.Error("List returned error")
	}

	if len(res.Volumes) != 1 {
		t.Error("Should have found 1 volume!")
	}

	createHelper(driver, t, name, mountpoint)

	res, err = driver.List()

	if err != nil {
		t.Error("List returned error2")
	}

	if len(res.Volumes) != 2 {
		t.Error("Should have found 1 volume!")
	}

	defaultCleanupHelper(driver, t)
	cleanupHelper(driver, t, name, mountpoint)
}

func TestMountUnmountPath(t *testing.T) {
	// TODO: Refactor this test. Is it okay to test all three at same time?
	driver := newLocalPersistDriver()

	defaultCreateHelper(driver, t)
	mountReq := &volume.MountRequest{Name: defaultTestName}
	unmountReq := &volume.UnmountRequest{Name: defaultTestName}
	pathReq := &volume.PathRequest{Name: defaultTestName}

	// mount, mount and path should have same output (they all use Path under the hood)

	mountRes, mountErr := driver.Mount(mountReq)
	unmountErr := driver.Unmount(unmountReq)
	pathRes, pathErr := driver.Path(pathReq)

	if !(mountErr == nil &&
		unmountErr == nil &&
		pathErr == nil) {
		t.Error("Error on mount, unmount or path")
	}

	if !(pathRes.Mountpoint == mountRes.Mountpoint &&
		pathRes.Mountpoint == defaultTestMountpoint) {
		t.Error("Mount, Unmount and Path should all return the same Mountpoint")
	}
}

func createHelper(driver localPersistDriver, t *testing.T, name string, mountpoint string) {

	req := &volume.CreateRequest{
		Name: name,
		Options: map[string]string{
			"mountpoint": mountpoint,
		},
	}

	err := driver.Create(req)

	if err != nil {
		t.Error(err)
	}
}

func defaultCreateHelper(driver localPersistDriver, t *testing.T) {
	createHelper(driver, t, defaultTestName, defaultTestMountpoint)
}

func cleanupHelper(driver localPersistDriver, t *testing.T, name string, mountpoint string) {
	os.RemoveAll(mountpoint)

	_, err := os.Stat(mountpoint)
	if !os.IsNotExist(err) {
		t.Error("[Cleanup] Mountpoint still exists:", err.Error())
	}

	removeReq := &volume.RemoveRequest{Name: name}

	err = driver.Remove(removeReq)
	if err != nil {
		t.Error("Remove failed", err)
	}

	getReq := &volume.GetRequest{Name: name}

	//Volume should be nil, as it is deleted
	v, err := driver.Get(getReq)

	if v != nil {
		t.Error(err)
	}
}

func defaultCleanupHelper(driver localPersistDriver, t *testing.T) {
	cleanupHelper(driver, t, defaultTestName, defaultTestMountpoint)
}
