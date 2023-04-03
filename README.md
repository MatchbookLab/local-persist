# Local Persist Volume Plugin for Docker

Fork of [local-persist](https://github.com/MatchbookLab/local-persist)

Create named local volumes that persist in the location(s) you want!

## Usage

1. Find the latest version in [Github releases](https://github.com/Carbonique/local-persist/releases) and find the corresponding image in [GHCR](https://github.com/Carbonique/local-persist/pkgs/container/local-persist)
2. Install the plugin (Use `docker plugin install` instead of `docker pull`. GHCR is unaware that `local-persist` is a docker plugin)

```sh
# Create the default directories for state.source and date.source (see below how to use different directories)
sudo mkdir -p /docker-plugins/local-persist/state /docker-plugins/local-persist/data

# to install
docker plugin install ghcr.io/carbonique/local-persist:<VERSION>-<ARCH> --alias=local-persist

# to enable debug
docker plugin install ghcr.io/carbonique/local-persist:<VERSION>-<ARCH> --alias=local-persist DEBUG=1

# or to change where plugin state is stored
docker plugin install ghcr.io/carbonique/local-persist:<VERSION>-<ARCH> --alias=local-persist state.source=<any_folder>

# or to change where volumes are stored
# the volumes will be created relative to the data.source directory; so the full path is data.source + mountpoint(args) if mountpoint option is provided
# else it will be data.source + volume name
docker plugin install ghcr.io/carbonique/local-persist:<VERSION>-<ARCH> --alias=local-persist data.source=<any_folder>
```

3. Create a volume
```sh
# to mount to directory data.source/test-mountpoint" (default: /docker-plugins/local-persist/data/test-mountpoint)
docker volume create -d local-persist -o mountpoint=test-mountpoint test-volume

# or without mountpoint option. The mountpoint will then be the name of the volume; e.g. data.source/test-volume (default: /docker-plugins/local-persist/data/test-volume)
docker volume create -d local-persist test-volume
```

## Goals of this fork:

1. Updating dependencies and using the new Docker driver interface
2. Implementing a V2 managed plugin, instead of a non-managed (legacy) plugin. A managed plugin makes life much easier, as there is no need for systemd and the plugin can simply be installed using `docker plugin install`
3. Multi-arch support and build using Github actions

## Rationale

The `local-persist` plugin gives you the same benefits of standalone Volumes that `docker volume create ...` normally affords, while also allowing you to create Volumes that *persist*, because your data *will not be deleted* when the Volume is removed. The `local` driver deletes all data when it's removed. With the `local-persist` driver, if you remove the driver, and then recreate it later with the same command above, any volume that was added to that volume will *still be there*.

Additionally the `local-persist` plugin allows you to store `docker volume` data wherever you want.

The above two goals could be achieved by using bind mounts, but these come with [drawbacks](#bind-mounts). The goal of storing data where you want to could be achieved using [named volumes with `driver_opts`](#named-volumes-with-driver_opts), but the data within these volumes will not persist upon deletion. And to be fair, I also made this fork to get some more exerience with Go.

### Bind mounts

Bind mounts come with several [drawbacks](https://docs.docker.com/storage/bind-mounts/#differences-between--v-and---mount-behavior) compared to named volumes. For example; Docker does not create the directories and the user will need to make sure that the container user has the correct permissions to access the mounted directories.

### Named volumes with driver_opts

One could mount named volumes to arbitrary directories in the following way:

```yml
version: '3'
services:
  db:
    image: mysql
    volumes:
      - dbdata:/var/lib/mysql
volumes:
  dbdata:
    driver: local
    driver_opts:
      type: 'none'
      o: 'bind'
      device: '/docker/sql'
```

A disadvantage of the method above compared to `local-persist` is that the `/docker/sql` directory needs to be created by the user.
And the contents of `/docker/sql` would not persist when deleted through `docker volume rm`

## Development

### Building

NOTE: the scripts below assume the user can run `docker` commands without needing `sudo`! If the user needs `sudo`, simply prepend the commands with `sudo` (`sudo ./scripts/build.sh`)

To run unit tests `go test ./driver`

To build run: `./scripts/build.sh <architecture>` (e.g.: `./scripts/build.sh amd64`)

To install run: `./scripts/install.sh <your-tag-for-the-plugin>` (e.g. `./scripts/install.sh local`)

To test run: `./scripts/integration_test.sh <your-tag-for-the-plugin>` (e.g. `./scripts/integration_test.sh local`)

To cleanup run: `./scripts/cleanup_plugin.sh <your-tag-for-the-plugin>` (e.g. `./scripts/cleanup_plugin.sh local`)

Or all in one go: `./scripts/build-install-integration_test-cleanup_plugin.sh` (e.g. `./scripts/build-install-integration_test-cleanup_plugin.sh`)

### Plugin configuration

Finding out how plugin configuration works was a bit of a puzzle, as the Docker plugin documentation is not great.

The following two things are important when configuring the plugin:

1. The `mounts` sections.
2. The `propagatedMount`.

```json
  "mounts": [
    {
      "description": "A place to store the plugin state so it can restore in between restarts. Source must be an existing path on host. Destination is path within the container.",
      "destination": "/local-persist/state",
      "options": [
        "rbind"
      ],
      "name": "state",
      "source": "/docker-plugins/local-persist/state",
      "settable": [
        "source"
      ],
      "type": "bind"
    },
    {
      "description": "A mount to share your data on. Source must be an existing path on host. Destination is path within the container.",
      "destination": "/local-persist/data",
      "options": [
        "rbind"
      ],
      "name": "data",
      "source": "/docker-plugins/local-persist/data",
      "settable": [
        "source"
      ],
      "type": "bind"
    }
  ],
  "propagatedMount": "/local-persist/data"
```

#### Mounts

The mounts section makes it possible to mount paths that reside inside the plugin rootfs to the host.
The path inside the plugin rootfs is the `destination` and the path on the Docker host is the `source`
This makes it possible to save the state and Docker volume outside of the plugin rootfs, so that these two will not be deleted on a plugin upgrade or on plugin deletion.

#### PropagedMount

The `propagatedMount` mounts the specified path within the plugin rootfs to make visible for the Docker daemon. If `propagedMount` would be omitted from the config, volume creation would still work. However, the daemon will not be able to mount the actual volumes, as it does not have access to the plugin rootfs. Making the volumes useless, as containers cannot mount them.