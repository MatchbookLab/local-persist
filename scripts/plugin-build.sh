#!/usr/bin/env bash

pwd

docker build -t local-persist-plugin .
ID=$(docker create local-persist-plugin true)
sudo rm -Rf plugin/rootfs
sudo mkdir plugin/rootfs
sudo docker export "$ID" | sudo tar -x -C plugin/rootfs
docker rm -vf "$ID"
docker rmi local-persist-plugin
