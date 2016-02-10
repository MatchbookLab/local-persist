#!/bin/bash

set -e

function create-volume {
    VOLUME=`docker volume create --driver=local-persist --opt mountpoint=/tmp/local-persist-integration/ --name=test-data`
}

function create-containers {
    ONE=`docker run -d -v test-data:/app/data/ alpine sleep 30`
    TWO=`docker run -d -v test-data:/src/data/ alpine sleep 30`
}

function check-containers {
    (docker exec $ONE cat /app/data/test.txt | grep 'Cameron Spear') || exit 111
    (docker exec $TWO cat /src/data/test.txt | grep 'Cameron Spear') || exit 222
}

function clean {
    docker rm -f $ONE
    docker rm -f $TWO
    docker volume rm $VOLUME
}

# setup
create-volume
create-containers

# copy a test file (note how this subtly breaks integration tests if my name is removed from the LICENSE ;-))
docker cp LICENSE $ONE:/app/data/test.txt

# check that the file exists in both
check-containers

# delete everything (start over point)
clean


# do it all again, but this time, DON'T manually copy a file... it should have persisted from before!
create-volume
create-containers

# if we were just using the `local` driver, this step would fail
check-containers

clean

echo -e "\nSuccess!"
