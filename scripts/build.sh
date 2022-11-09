mkdir -p ./plugin/rootfs
rm -rf ./plugin/rootfs/*
docker build -t rootfsimage .
id=$(docker create rootfsimage true) 
docker export "$id" | sudo tar -x -C ./plugin/rootfs

