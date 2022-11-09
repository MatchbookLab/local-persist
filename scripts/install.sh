TAG=$1
PLUGIN=ghcr.io/carbonique/local-persist:${TAG}

docker plugin disable ${PLUGIN}
docker plugin rm ${PLUGIN}
docker plugin create ${PLUGIN} ./plugin
docker plugin enable ${PLUGIN}

