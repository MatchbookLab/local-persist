package driver

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/docker/go-plugins-helpers/volume"
)

const (
	defaultTestName       = "test-volume"
	defaultTestMountpoint = "./test/data/local-persist-test"
	stateDir              = "./test/state"
	dataDir               = "./test/data"
)

func init() {
	if _, err := os.Stat(defaultTestMountpoint); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(defaultTestMountpoint, os.ModePerm)
		if err != nil {
			os.Exit(1)
		}
	}
}

func TestCreate(t *testing.T) {
	driver := createDriverHelper()

	defaultCreateVolumeHelper(driver, t)

	// test that a directory is created
	_, err := os.Stat(defaultTestMountpoint)
	if os.IsNotExist(err) {
		t.Error("Mountpoint directory was not created:", err.Error())
	}

	// test that volumes has one
	if len(driver.volumes) != 1 {
		t.Error("Driver should have exactly 1 volume")
	}

	defaultVolumeCleanupHelper(driver, t)

	req := &volume.CreateRequest{Name: "defaultTestName"}
	// test that options are required
	err = driver.Create(req)

	if err.Error() != "the `mountpoint` option is required" {
		t.Error(err)
	}
	defaultVolumeCleanupHelper(driver, t)
}

func TestGet(t *testing.T) {
	driver := createDriverHelper()

	defaultCreateVolumeHelper(driver, t)

	req := &volume.GetRequest{Name: defaultTestName}

	res, err := driver.Get(req)
	if err != nil {
		t.Error("Should have found a volume!")
	}

	if res.Volume.Name != defaultTestName {
		t.Error("Incorrect volume name was returned")
	}

	defaultVolumeCleanupHelper(driver, t)
}

func TestList(t *testing.T) {
	driver := createDriverHelper()

	name := defaultTestName + "2"
	mountpoint := defaultTestMountpoint + "2"

	defaultCreateVolumeHelper(driver, t)

	res, err := driver.List()

	if err != nil {
		t.Error("List returned error")
	}

	if len(res.Volumes) != 1 {
		t.Error("Should have found 1 volume!")
	}

	createVolumeHelper(driver, t, name, mountpoint)

	res, err = driver.List()

	if err != nil {
		t.Error("List returned error2")
	}

	if len(res.Volumes) != 2 {
		t.Error("Should have found 1 volume!")
	}

	defaultVolumeCleanupHelper(driver, t)
	cleanupVolumeHelper(driver, t, name, mountpoint)
}

func TestMount(t *testing.T) {
	driver := createDriverHelper()

	defaultCreateVolumeHelper(driver, t)

	req := &volume.MountRequest{Name: defaultTestName}
	_, err := driver.Mount(req)

	if err != nil {
		t.Error("Error on mount")
	}

	// Remove a mountpoint, while volume still exists
	err = os.Remove(defaultTestMountpoint)

	if err != nil {
		t.Error("Could not remove mountpoint")
	}

	_, err = driver.Mount(req)
	if err == nil {
		t.Error("Mountpoint was deleted but test did not error")
	}

	// Test to mount a existing file (should not be possible)
	_, err = os.Create(defaultTestMountpoint)
	if err != nil {
		t.Error("Could not create mountpoint as file")
	}
	_, err = driver.Mount(req)
	if err == nil {
		t.Error("Mountpoint is a file but test did not error")
	}
	defaultVolumeCleanupHelper(driver, t)
}

func TestUnmount(t *testing.T) {
	driver := createDriverHelper()

	defaultCreateVolumeHelper(driver, t)

	// Requesting an existing volume
	req := &volume.UnmountRequest{Name: defaultTestName}

	err := driver.Unmount(req)

	if err != nil {
		t.Error("Error on unmount")
	}

	// Requesting a non-existing volume
	reqFail := &volume.UnmountRequest{Name: defaultTestName + "does_not_exist"}
	err = driver.Unmount(reqFail)

	if err == nil {
		t.Error("Test should fail as volume does not exist")
	}

	defaultVolumeCleanupHelper(driver, t)

}
func TestPath(t *testing.T) {
	driver := createDriverHelper()

	defaultCreateVolumeHelper(driver, t)
	// Requesting an existing volume
	req := &volume.PathRequest{Name: defaultTestName}

	v, err := driver.Path(req)

	if err != nil {
		t.Error("Error on path")
	}

	if v.Mountpoint != defaultTestMountpoint {
		t.Error("Mountpoint should be equal to defaultTestMountpoint")
	}

	// Requesting a non-existing volume
	reqFail := &volume.PathRequest{Name: defaultTestName + "does_not_exist"}
	_, err = driver.Path(reqFail)

	if err == nil {
		t.Error("Test should fail as volume does not exist")
	}

	defaultVolumeCleanupHelper(driver, t)

}

func createVolumeHelper(driver *localPersistDriver, t *testing.T, name string, mountpoint string) {

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

func defaultCreateVolumeHelper(driver *localPersistDriver, t *testing.T) {
	createVolumeHelper(driver, t, defaultTestName, defaultTestMountpoint)
}

func cleanupVolumeHelper(driver *localPersistDriver, t *testing.T, name string, mountpoint string) {
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

	if v.Volume != nil {
		t.Error(err)
	}
}

func defaultVolumeCleanupHelper(driver *localPersistDriver, t *testing.T) {
	cleanupVolumeHelper(driver, t, defaultTestName, defaultTestMountpoint)
}

func createDriverHelper() *localPersistDriver {
	d, err := NewLocalPersistDriver(stateDir, dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	return d
}
