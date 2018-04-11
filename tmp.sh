#!/usr/bin/env bash

export PLUGIN_NAME=local-persist

docker build -t ${PLUGIN_NAME} .

docker build -t ${PLUGIN_NAME}:rootfs .
mkdir -p ./plugin/rootfs
docker create --name tmp ${PLUGIN_NAME}:rootfs
docker export tmp | tar -x -C ./plugin/rootfs
cp -f config.json ./plugin/
docker rm -vf tmp

docker plugin rm -f ${PLUGIN_NAME} || true
docker plugin create ${PLUGIN_NAME} ./plugin
docker plugin enable ${PLUGIN_NAME}
