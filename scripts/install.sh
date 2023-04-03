TAG=$1
PLUGIN=ghcr.io/carbonique/local-persist:${TAG}
STATE_DIR="$(pwd)/test/state"
DATA_DIR="$(pwd)/test/data"

mkdir -p $STATE_DIR
mkdir -p $DATA_DIR

docker plugin disable ${PLUGIN}
docker plugin rm ${PLUGIN}
docker plugin create ${PLUGIN} ./plugin
docker plugin set ${PLUGIN} data.source=$DATA_DIR state.source=$STATE_DIR
docker plugin enable ${PLUGIN}

