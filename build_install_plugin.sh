mkdir -p ./plugin/rootfs
rm -rf ./plugin/rootfs/*
docker build -t rootfsimage .
id=$(docker create rootfsimage true) 
docker export "$id" | sudo tar -x -C ./plugin/rootfs

docker plugin disable local-persist
docker plugin rm local-persist
docker plugin create local-persist ./plugin
docker plugin enable local-persist

