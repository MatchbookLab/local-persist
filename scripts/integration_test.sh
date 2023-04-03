#!/bin/bash

set -e

TAG=$1
PLUGIN=ghcr.io/carbonique/local-persist:${TAG}

function create-volume {
  printf "\nCreating volume \n"
  VOLUME=`docker volume create --driver=${PLUGIN} --name=test-data $@`
}

function create-containers {
  printf "\nCreating containers \n"
  ONE=`docker run -d -v test-data:/app/data/ alpine sleep 30`
  TWO=`docker run -d -v test-data:/src/data/ alpine sleep 30`
}

function check-containers {
  printf "\nChecking containers \n"
  (docker exec $ONE cat /app/data/test.txt | grep 'Cameron Spear') || exit 111
  (docker exec $TWO cat /src/data/test.txt | grep 'Cameron Spear') || exit 222
}

function clean {
  printf "\nClean \n"
  docker rm -f $ONE
  docker rm -f $TWO
  docker volume rm $VOLUME
}

function test {
  printf "\nRunning test \n"
  create-volume $1
  create-containers

  # copy a test file (note how this subtly breaks integration tests if my name is removed from the LICENSE ;-))
  docker cp LICENSE $ONE:/app/data/test.txt

  # check that the file exists in both
  check-containers

  # delete everything (start over point)
  clean

  # do it all again, but this time, DON'T manually copy a file... it should have persisted from before!
  create-volume $1
  create-containers

  # if we were just using the `local` driver, this step would fail
  check-containers

  clean
}

mkdir -p ./test
# Run test with mountpoint
test "--opt mountpoint=/local-persist-integration"

printf "\nSecond run without mountpoint option"

# run test without mountpoint
test 

echo -e "\nSuccess!"
