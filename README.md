# Local Persist Volume Plugin for Docker

Fork of [local-persist](https://github.com/MatchbookLab/local-persist)

Create named local volumes that persist in the location(s) you want!

Goals of this fork:

1. Updated dependencies + updated Docker driver interface
2. Build using Github actions
3. Implement a V2 managed plugin, instead of a non-managed (legacy) plugin. A managed plugin makes life much easier, as there is no need for systemd and the plugin can simply be installed using `docker plugin install` However, it also comes with restrictions, as the data can now only be stored in the `/docker-data` directory

## Rationale

In Docker 1.9, they added support for [creating standalone named Volumes](https://docs.docker.com/engine/reference/commandline/volume_create/). Now with Docker 1.10 and Docker Compose 1.6's new syntax, you can [create named volumes through Docker Compose](https://docs.docker.com/compose/compose-file/#volume-configuration-reference).

This is great for creating standalone volumes and easily connecting them to different directories in different containers as a way to share data between multiple containers. On a much larger scale, it also allows for the use of Docker Volume Plugins to do cool things like [Flocker](https://github.com/ClusterHQ/flocker) is doing (help run stateful containers across multiple hosts).

Even if something like Flocker is overkill for your needs, it can still be useful to have persistent data on your host. I'm a strong advocate for "Docker for small projects" and not just huge, scaling behemoths and microservices. I wrote this out of a need on projects I'm currently working on and have in production.

This `local-persist` approach gives you the same benefits of standalone Volumes that `docker volume create ...` normally affords, while also allowing you to create Volumes that *persist*, thus giving those stateful containers their state. Read below how to install and use, then read more about the [benefits](#benefits) of this approach.

## Installing & Running

Docker Engine’s plugin system allows you to install, start, stop, and remove plugins using Docker Engine.

The plugin can be installed using `docker plugin install`

## Usage: Creating Volumes

Then to use, you can create a volume with this plugin (this example will be for a shared folder for images):

```shell
docker volume create -d local-persist -o mountpoint=/docker-data/images --name=images
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