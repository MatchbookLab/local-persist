#!/usr/bin/env bash

LOCAL_PERSIST=$(docker build -q -f Dockerfile-build .)
docker run -it -v `pwd`/bin:/go/src/local-persist/bin $LOCAL_PERSIST
