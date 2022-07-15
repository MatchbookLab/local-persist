mkdir -p ./plugin/rootfs
docker build -t rootfsimage .
id=$(docker create rootfsimage true) 
sudo docker export "$id" | sudo tar -x -C ./plugin/rootfs

docker plugin create test-persist ./plugin