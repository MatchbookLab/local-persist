TAG=$1
PLUGIN=ghrc.io/carbonique/local-persist:${TAG}

docker plugin disable ${PLUGIN}
docker plugin rm ${PLUGIN}