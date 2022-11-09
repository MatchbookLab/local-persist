
# This directory will on purpose not be removed after the test.
# It could be that its used for storing 'real' data as well
mkdir -p /docker-data

./scripts/build.sh
./scripts/install.sh "local"
./scripts/integration_test.sh "local"
./scripts/cleanup_plugin.sh "local"