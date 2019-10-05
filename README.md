# Local Persist Volume Plugin for Docker

[![Build Status](https://travis-ci.org/MatchbookLab/local-persist.svg?branch=master)](https://travis-ci.org/MatchbookLab/local-persist) [![Join the chat at https://gitter.im/MatchbookLab/local-persist](https://badges.gitter.im/MatchbookLab/local-persist.svg)](https://gitter.im/MatchbookLab/local-persist?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Create named local volumes that persist in the location(s) you want!

## Rationale

In Docker 1.9, they added support for [creating standalone named Volumes](https://docs.docker.com/engine/reference/commandline/volume_create/). Now with Docker 1.10 and Docker Compose 1.6's new syntax, you can [create named volumes through Docker Compose](https://docs.docker.com/compose/compose-file/#volume-configuration-reference).

This is great for creating standalone volumes and easily connecting them to different directories in different containers as a way to share data between multiple containers. On a much larger scale, it also allows for the use of Docker Volume Plugins to do cool things like [Flocker](https://github.com/ClusterHQ/flocker) is doing (help run stateful containers across multiple hosts).

Even if something like Flocker is overkill for your needs, it can still be useful to have persistent data on your host. I'm a strong advocate for "Docker for small projects" and not just huge, scaling behemoths and microservices. I wrote this out of a need on projects I'm currently working on and have in production.

This `local-persist` approach gives you the same benefits of standalone Volumes that `docker volume create ...` normally affords, while also allowing you to create Volumes that *persist*, thus giving those stateful containers their state. Read below how to install and use, then read more about the [benefits](#benefits) of this approach.

## Installing & Running

To create a Docker Plugin, one must create a Unix socket that Docker will look for when you use the plugin and then it listens for commands from Docker and runs the corresponding code when necessary.

Running the code in this project with create the said socket, listening for commands from Docker to create the necessary Volumes.

According to the [Docker Plugin API Docs](https://docs.docker.com/engine/extend/plugin_api/):

> Plugins can run inside or outside containers. Currently running them outside containers is recommended.

It doesn't really say *why* one way is recommended over the other, but I provide binaries and instructions to run outside of container, as well as an image and instructions to run it inside a container.

### Running Outside a Container

**Note:** You currently cannot run this plugin natively on macOS or Windows. The current workaround is to [run the plugin in a container](#running-from-within-a-container).

#### Quick Way

I provide an `install` script that will download the proper binary, set up an Systemd service to start when Docker does and enable it.

```shell
curl -fsSL https://raw.githubusercontent.com/MatchbookLab/local-persist/master/scripts/install.sh | sudo bash
```

This needs be to run on the Docker *host*. i.e. running that on a Mac won't work (and it will print a message saying as much and exit).

This has been tested on Ubuntu 15.10, and is known *not* to work on CoreOS (yet). If you need to use Upstart instead of Systemd, you can pass the `--upstart` flag to the install script, but it isn't as tested, so it may not work:

```shell
curl -fsSL https://raw.githubusercontent.com/MatchbookLab/local-persist/master/scripts/install.sh | sudo bash -s -- --upstart
```

Follow the same process to update to the latest version.

#### Manual Way

If you're uncomfortable running a script you downloaded off the internet with `sudo`, you can extract any of the steps out of the [`install.sh`](scripts/install.sh) script and run them manually. However you want to do it, the main steps are:

1. Download the appropriate binary from the [Releases page](https://github.com/MatchbookLab/local-persist/releases) for your OS and architecture.
2. Rename the downloaded file `docker-volume-local-persist`
3. Place it in `/usr/bin` (you can put it somewhere else, but be sure your Systemd (or similar) config reflects the change).
4. Make sure the file is executable (`chmod +x /usr/bin/docker-volume-local-persist`)
5. It's enough to just run it at this point (type `docker-volume-local-persist` and hit enter) to test, etc, and if that's all you're trying to do, you're done. But if you want it to start with Docker, proceed to step 6.
6. Download [systemd.service](init/systemd.service)
7. Rename the service file to `docker-volume-local-persist.service`
8. Move it to `/etc/systemd/system/`
9. run `sudo systemctl daemon-reload` to reload the config
10. run `sudo systemctl enable docker-volume-local-persist` to enable the service (it will start after Docker does)
11. run `sudo systemctl start docker-volume-local-persist` to start it now. Safe to run if it's already started

<a id="running-from-within-a-container"></a>
### Running from Within a Container (aka Running on Mac or Windows)

macOS and Windows do not support native Docker plugins, so the solution is to run this plugin from another container (you can also do this on Linux if you don't want to install the plugin manually).

I maintain an [image on Docker Hub](https://hub.docker.com/r/cwspear/docker-local-persist-volume-plugin/) to run this plugin from a container:

```shell
docker run -d \
    -v /run/docker/plugins/:/run/docker/plugins/ \
    -v /path/to/store/json/for/restart/:/var/lib/docker/plugin-data/ \
    -v /path/to/where/you/want/data/volume/:/path/to/where/you/want/data/volume/ \
        cwspear/docker-local-persist-volume-plugin
```

The `-v /run/docker/plugins/:/run/docker/plugins/` part will make sure the `sock` file gets created at the right place. You also need to add one or more volumes to places you want to mount your volumes later at.

For example, if I am going to persist my MySQL data for a container I'm going to build later at `/data/mysql/`, I would add a `-v /data/mysql/:/data/mysql/` to the command above (or even `-v /data/:/data/`). You can add more than one location in this manner.

Lastly, the `-v /path/to/store/json/for/restart/:/var/lib/docker/plugin-data/` part is so that the plugin can create a `json` file to know what volumes existed in case of a system restart, etc. 

When the container is destroyed, etc, it will look at a file it created in `/var/lib/docker/plugin-data/` to recreate any volumes that had previously existed, so you want that JSON file to persist on the host. 

## Usage: Creating Volumes

Then to use, you can create a volume with this plugin (this example will be for a shared folder for images):

```shell
docker volume create -d local-persist -o mountpoint=/data/images --name=images
```

Then if you create a container, you can connect it to this Volume:

```shell
docker run -d -v images:/path/to/images/on/one/ one
docker run -d -v images:/path/to/images/on/two/ two
# etc
```

Also, see [docker-compose.example.yml](docker-compose.example.yml) for an example to do something like this with Docker Compose (needs Compose 1.6+ which needs Engine 1.10+).

## Benefits

This has a few advantages over the (default) `local` driver that comes with Docker, because our data *will not be deleted* when the Volume is removed. The `local` driver deletes all data when it's removed. With the `local-persist` driver, if you remove the driver, and then recreate it later with the same command above, any volume that was added to that volume will *still be there*.

You may have noticed that you could do this with data-only containers, too. And that's true, and using that technique has a few advantages, one thing it (specifically as a limitation of `volumes-from`) does *not* allow, is mounting that shared volume to a different path inside your containers. Trying to recreate the above example, each container would have to store images in the same directory in their containers, instead of separate ones which `local-persist` allows.

Also, using `local-persist` instead of data-only containers, `docker ps -a` won't have extra dead entries, and `docker volume ls` will have more descriptive output (because volumes have names).
