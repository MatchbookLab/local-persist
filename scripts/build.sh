mkdir -p ./plugin/rootfs
rm -rf ./plugin/rootfs/*
docker buildx build --platform=$1 -t rootfsimage .
id=$(docker create --platform=$1 rootfsimage true) 
docker export "$id" | sudo tar -x -C ./plugin/rootfs

