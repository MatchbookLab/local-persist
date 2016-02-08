# Local Persist Volume Plugin for Docker

[![Build Status](https://travis-ci.org/CWSpear/docker-local-persist-volume-plugin.svg?branch=master)](https://travis-ci.org/CWSpear/docker-local-persist-volume-plugin)

Create named local volumes that persist in the location(s) you want!

## Rationale 

In Docker 1.9, they added support for [creating standalone named Volumes](https://docs.docker.com/engine/reference/commandline/volume_create/). Now with Docker 1.10 and Docker Compose 1.6's new syntax, you can [create named volumes through Docker Compose](https://docs.docker.com/compose/compose-file/#volume-configuration-reference:91de898b5f5cdb090642a917d3dedf68).
 
 This is great for creating standalone volumes and easily connecting them to different directories in different containers as a way to share data between multiple containers. On a much larger scale, it also allows for the use of Docker Volume Plugins to do cool things like [Flocker](https://github.com/ClusterHQ/flocker) is doing (help run stateful containers across multiple hosts).
 
If something like Flocker is overkill for your needs, it can still be useful for

## Installing & Running

To create a Docker Plugin, one must create a Unix socket that Docker will look for when you use the plugin and then it listens for commands from Docker and runs the corresponding code when necessary.

Running the code in this project with create the said socket, listening for commands from Docker to create the necessary Volumes.

Pre-1.0, you can either build and run the project locally, or use this with Docker (run this on your Docker host):

```shell
docker run -d \
    -v /run/docker/plugins/:/run/docker/plugins/ \
    -v /path/to/where/you/want/data/volume/:/path/to/where/you/want/data/volume/ \ 
        cwspear/docker-local-persist-volume-plugin
```

The `-v /run/docker/plugins/:/run/docker/plugins/` part will make sure the `sock` file gets created at the right place. You also need to add one or more volumes to places you want to mount your volumes later at.

For example, if I am going to persist my MySQL data for a container I'm going to build later at `/data/mysql/`, I would add a `-v /data/mysql/:/data/mysql/` to the command above (or even `-v /data/:/data/`).

## Usage

Then to use, you can create a volume with this plugin (this example will be for a shared folder for images):

```shell
docker create volume create -d local-persist -o mountpoint=/data/images --name=images
```

Then if you create a container, you can connect it to this Volume:

```shell
docker run -d -v images:/path/to/images/on/one/ one
docker run -d -v images:/path/to/images/on/two/ two
# etc
```

Also, see [docker-compose.example.yml](https://github.com/CWSpear/docker-local-persist-volume-plugin/blob/master/docker-compose.example.yml) for an example to do something like this with Docker Compose (needs Compose 1.6+ and Engine 1.10+).

## Benefits

This has a few advantages over the (default) `local` driver that comes with Docker, because our data *will not be deleted* when the Volume is removed. The `local` driver deletes all data when it's removed. With the `local-persist` driver, if you remove the driver, and then recreate it later with the same command above, any volume that was added to that volume will *still be there*.

You may have noticed that you could do this with data-only containers, too. And that's true, and using that technique has a few advantages, one thing it (specifically as a limitation of `volumes-from`) does *not* allow, is mounting that shared volume to a different path inside your containers. Trying to recreate the above example, each container would have to store images in the same directory in their containers, instead of separate ones which `local-persist` allows.

Also, using `local-persist` instead of data-only containers, `docker ps -a` won't have extra dead entries, and `docker volume ls` will have more descriptive output (because volumes have names).

