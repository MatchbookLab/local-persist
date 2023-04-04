package driver

import (
	"os"
	"path"
	"reflect"
	"sync"
	"testing"

	"github.com/docker/go-plugins-helpers/volume"
)

const (
	BASEDIR       = "./test"
	DATAPATH      = "./test/data"
	STATEPATH     = "./test/state"
	STATEFILEPATH = "./test/state/test-local-persist.json"
)

var volume1 = volume.Volume{
	Name:       "test-volume-1",
	Mountpoint: path.Join("myDir", "test-volume-1"),
}

var volume2 = volume.Volume{
	Name:       "test-volume-2",
	Mountpoint: path.Join("test-volume-2"),
}

type fields struct {
	Name          string
	volumes       map[string]string
	mutex         *sync.Mutex
	stateFilePath string
	dataPath      string
}

func returnFieldsEmptyVolume() fields {
	vol := make(map[string]string)

	f := fields{
		Name:          "local-persist-test",
		volumes:       vol,
		mutex:         &sync.Mutex{},
		stateFilePath: STATEFILEPATH,
		dataPath:      DATAPATH,
	}
	return f
}

func returnFieldsOneVolume() fields {
	vol := make(map[string]string)

	vol[volume1.Name] = volume1.Mountpoint

	f := fields{
		Name:          "local-persist-test",
		volumes:       vol,
		mutex:         &sync.Mutex{},
		stateFilePath: STATEFILEPATH,
		dataPath:      DATAPATH,
	}
	return f
}

func returnFieldsTwoVolumes() fields {
	vol := make(map[string]string)

	vol[volume1.Name] = volume1.Mountpoint
	vol[volume2.Name] = volume2.Mountpoint

	f := fields{
		Name:          "local-persist-test",
		volumes:       vol,
		mutex:         &sync.Mutex{},
		stateFilePath: STATEFILEPATH,
		dataPath:      DATAPATH,
	}
	return f
}

func cleanupBaseDir() {
	os.RemoveAll(path.Join(BASEDIR, "/"))
}

func setupDirs(t *testing.T) {
	err := ensureDir(DATAPATH, 0700)
	if err != nil {
		t.Errorf("Could not ensureDir %s", DATAPATH)
	}

	err = ensureDir(STATEPATH, 0700)
	if err != nil {
		t.Errorf("Could not ensureDir %s", STATEPATH)
	}
}

func Test_localPersistDriver_Get(t *testing.T) {

	type args struct {
		req *volume.GetRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *volume.GetResponse
		wantErr bool
	}{
		{
			name:   "Get volume, should pass",
			fields: returnFieldsOneVolume(),
			args: args{&volume.GetRequest{
				Name: volume1.Name,
			}},
			want: &volume.GetResponse{
				Volume: &volume1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &localPersistDriver{
				Name:          tt.fields.Name,
				volumes:       tt.fields.volumes,
				mutex:         tt.fields.mutex,
				stateFilePath: tt.fields.stateFilePath,
				dataPath:      tt.fields.dataPath,
			}
			got, err := driver.Get(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("localPersistDriver.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("localPersistDriver.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_localPersistDriver_List(t *testing.T) {

	tests := []struct {
		name    string
		fields  fields
		want    *volume.ListResponse
		wantErr bool
	}{
		{
			name:    "List empty volumes, should pass",
			fields:  returnFieldsEmptyVolume(),
			want:    &volume.ListResponse{},
			wantErr: false,
		},
		{
			name:   "List one volume, should pass",
			fields: returnFieldsOneVolume(),
			want: &volume.ListResponse{
				Volumes: []*volume.Volume{&volume1},
			},
			wantErr: false,
		},
		{
			name:   "List two volumes, should pass",
			fields: returnFieldsTwoVolumes(),
			want: &volume.ListResponse{
				Volumes: []*volume.Volume{&volume1, &volume2},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &localPersistDriver{
				Name:          tt.fields.Name,
				volumes:       tt.fields.volumes,
				mutex:         tt.fields.mutex,
				stateFilePath: tt.fields.stateFilePath,
				dataPath:      tt.fields.dataPath,
			}
			got, err := driver.List()
			if (err != nil) != tt.wantErr {
				t.Errorf("localPersistDriver.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("localPersistDriver.List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_localPersistDriver_Create(t *testing.T) {
	setupDirs(t)
	defer cleanupBaseDir()

	custom_mountpoint_option := make(map[string]string)
	custom_mountpoint_option["mountpoint"] = "my-random-mountpoint"
	type directory struct {
		path string
	}
	type args struct {
		req *volume.CreateRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    directory
		wantErr bool
	}{
		{
			name:   "Create volume with mountpoint option, should pass",
			fields: returnFieldsEmptyVolume(),
			args: args{
				req: &volume.CreateRequest{
					Name:    volume1.Name,
					Options: custom_mountpoint_option,
				},
			},
			want: directory{
				path: path.Join(DATAPATH, custom_mountpoint_option["mountpoint"]),
			},
			wantErr: false,
		},
		{
			name:   "Create volume without mountpoint option, should pass",
			fields: returnFieldsEmptyVolume(),
			args: args{
				&volume.CreateRequest{
					Name: volume2.Name,
				},
			},
			want: directory{
				path: path.Join(DATAPATH, volume2.Name),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &localPersistDriver{
				Name:          tt.fields.Name,
				volumes:       tt.fields.volumes,
				mutex:         tt.fields.mutex,
				stateFilePath: tt.fields.stateFilePath,
				dataPath:      tt.fields.dataPath,
			}
			if err := driver.Create(tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("localPersistDriver.Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if _, err := os.Stat(tt.want.path); os.IsNotExist(err) && !tt.wantErr {
				t.Errorf("localPersistDriver.Create() error directory %s does not exist", tt.want.path)
			}
		})
	}
}

func Test_localPersistDriver_Remove(t *testing.T) {
	setupDirs(t)
	defer cleanupBaseDir()

	type args struct {
		req *volume.RemoveRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Remove existing volume, should pass",
			fields: returnFieldsTwoVolumes(),
			args: args{
				req: &volume.RemoveRequest{
					Name: volume1.Name,
				},
			},
			wantErr: false,
		},
		{
			name:   "Remove non-existing volume, should give error",
			fields: returnFieldsEmptyVolume(),
			args: args{
				req: &volume.RemoveRequest{
					Name: "i-do-not-exist",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &localPersistDriver{
				Name:          tt.fields.Name,
				volumes:       tt.fields.volumes,
				mutex:         tt.fields.mutex,
				stateFilePath: tt.fields.stateFilePath,
				dataPath:      tt.fields.dataPath,
			}
			if err := driver.Remove(tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("localPersistDriver.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_localPersistDriver_Mount(t *testing.T) {
	// TODO cleanup (Third volume needs the mountpoint including the base dir)
	// and it woulde be nice to add unique setup steps for each scenario
	setupDirs(t)
	defer cleanupBaseDir()

	// Create everything for mounting an existing volume and mountpoint
	// Existingvolume is needed as we need a volume that includes the mountpoint including the DATAPATH
	existingVolume := volume.Volume{
		Name:       "test-volume-3",
		Mountpoint: path.Join(DATAPATH, "test-volume-3"),
	}

	volumes := make(map[string]string)
	volumes[existingVolume.Name] = existingVolume.Mountpoint

	existingVolumeFields := fields{
		Name:          "local-persist-test",
		volumes:       volumes,
		mutex:         &sync.Mutex{},
		stateFilePath: STATEFILEPATH,
		dataPath:      DATAPATH,
	}

	err := ensureDir(existingVolume.Mountpoint, 0755)
	if err != nil {
		t.Errorf("Could not ensureDir %s", DATAPATH)
	}

	fileDisguisedAsVolume := volume.Volume{
		Name:       "file-volume",
		Mountpoint: path.Join(DATAPATH, "file-volume"),
	}

	// Create the file needed to test the mounting of a file
	volumes = make(map[string]string)
	volumes[fileDisguisedAsVolume.Name] = fileDisguisedAsVolume.Mountpoint

	fileDisguisedAsVolumeFields := fields{
		Name:          "local-persist-test",
		volumes:       volumes,
		mutex:         &sync.Mutex{},
		stateFilePath: STATEFILEPATH,
		dataPath:      DATAPATH,
	}

	_, err = os.Create(fileDisguisedAsVolume.Mountpoint)
	if err != nil {
		t.Errorf("Could not create file %s", fileDisguisedAsVolume.Mountpoint)
	}

	type args struct {
		req *volume.MountRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *volume.MountResponse
		wantErr bool
	}{
		{
			name:   "Mount an existing volume, should pass",
			fields: existingVolumeFields,
			args: args{
				&volume.MountRequest{
					Name: existingVolume.Name,
				},
			},
			want: &volume.MountResponse{
				Mountpoint: existingVolume.Mountpoint,
			},
			wantErr: false,
		},
		{
			name:   "Mount a non-existing volume, should give error",
			fields: returnFieldsOneVolume(),
			args: args{
				&volume.MountRequest{
					Name: volume1.Name,
				},
			},
			want:    &volume.MountResponse{},
			wantErr: true,
		},
		{
			name:   "Mount a non-existing directory, should give error",
			fields: returnFieldsOneVolume(),
			args: args{
				&volume.MountRequest{
					Name: volume1.Name,
				},
			},
			want:    &volume.MountResponse{},
			wantErr: true,
		},
		{
			name:   "Mount a file, should give error",
			fields: fileDisguisedAsVolumeFields,
			args: args{
				&volume.MountRequest{
					Name: fileDisguisedAsVolume.Name,
				},
			},
			want:    &volume.MountResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &localPersistDriver{
				Name:          tt.fields.Name,
				volumes:       tt.fields.volumes,
				mutex:         tt.fields.mutex,
				stateFilePath: tt.fields.stateFilePath,
				dataPath:      tt.fields.dataPath,
			}
			got, err := driver.Mount(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("localPersistDriver.Mount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("localPersistDriver.Mount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_localPersistDriver_Path(t *testing.T) {

	type args struct {
		req *volume.PathRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *volume.PathResponse
		wantErr bool
	}{
		{
			name:   "Request existing volume, should pass.",
			fields: returnFieldsOneVolume(),
			args: args{
				&volume.PathRequest{
					Name: volume1.Name,
				},
			},
			want: &volume.PathResponse{
				Mountpoint: volume1.Mountpoint,
			},
			wantErr: false,
		},
		{
			name:   "Request non-existing volume, should return error.",
			fields: returnFieldsEmptyVolume(),
			args: args{
				&volume.PathRequest{
					Name: volume1.Name,
				},
			},
			want:    &volume.PathResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &localPersistDriver{
				Name:          tt.fields.Name,
				volumes:       tt.fields.volumes,
				mutex:         tt.fields.mutex,
				stateFilePath: tt.fields.stateFilePath,
				dataPath:      tt.fields.dataPath,
			}
			got, err := driver.Path(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("localPersistDriver.Path() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("localPersistDriver.Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_localPersistDriver_Unmount(t *testing.T) {

	type args struct {
		req *volume.UnmountRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Request unmount of existing volume, should pass.",
			fields: returnFieldsOneVolume(),
			args: args{
				&volume.UnmountRequest{
					Name: volume1.Name,
				},
			},
			wantErr: false,
		},
		{
			name:   "Unmount non-existing volume, should give error.",
			fields: returnFieldsEmptyVolume(),
			args: args{
				&volume.UnmountRequest{
					Name: "unmount-this-non-existing-volume-please",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &localPersistDriver{
				Name:          tt.fields.Name,
				volumes:       tt.fields.volumes,
				mutex:         tt.fields.mutex,
				stateFilePath: tt.fields.stateFilePath,
				dataPath:      tt.fields.dataPath,
			}
			if err := driver.Unmount(tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("localPersistDriver.Unmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
