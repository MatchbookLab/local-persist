#!/usr/bin/env bash

black='\033[0;30m'        # Black
red='\033[0;31m'          # Red
green='\033[0;32m'        # Green
yellow='\033[0;33m'       # Yellow
blue='\033[0;34m'         # Blue
purple='\033[0;35m'       # Purple
cyan='\033[0;36m'         # Cyan
white='\033[0;37m'        # White
nocolor='\033[0m'         # No Color


USER=cwspear
REPO=local-persist

# check to make sure github-release is installed!
github-release --version > /dev/null || exit

if [[ $RELEASE_NAME == "" ]]; then
    echo -e ${cyan}Enter release name:${nocolor}
    read RELEASE_NAME
    echo ''
fi

if [[ $RELEASE_DESCRIPTION == "" ]]; then
    echo -e ${cyan}Enter release description:${nocolor}
    read RELEASE_DESCRIPTION
    echo ''
fi

if [[ $RELEASE_TAG == "" ]]; then
    printf "${cyan}Enter release tag:${nocolor} v"
    read RELEASE_TAG
    echo ''

    RELEASE_TAG="v${RELEASE_TAG}"
fi

if [[ $PRERELEASE == "" ]]; then
    printf "${cyan}Is this a prerelease? [yN]${nocolor} "
    read PRERELEASE_REPLY
    echo ''

    FIRST=${PRERELEASE_REPLY:0:1}
    echo $FIRST

    PRERELEASE=false
    [[ $FIRST == 'Y' || $FIRST == 'y' ]] && PRERELEASE=true
fi

sed -i "s|VERSION=\".*\"|VERSION=\"${RELEASE_TAG}\"|" scripts/install.sh
sed -i "s|ENV VERSION .*|ENV VERSION ${RELEASE_TAG}|" Dockerfile

git commit -am "Tagged ${RELEASE_TAG}"
git push
git tag -a $RELEASE_TAG -m "$RELEASE_NAME"
git push --tags

echo ''
echo Releasing...
echo ''
echo USER=$USER
echo REPO=$REPO
echo RELEASE_NAME="'$RELEASE_NAME'"
echo RELEASE_DESCRIPTION="'$RELEASE_DESCRIPTION'"
echo RELEASE_TAG=$RELEASE_TAG
echo PRERELEASE=$PRERELEASE
echo ''

if [[ "$PRERELEASE" == true ]]; then
    github-release release \
        --user $USER \
        --repo $REPO \
        --tag $RELEASE_TAG \
        --name "$RELEASE_NAME" \
        --description "$RELEASE_DESCRIPTION" \
        --pre-release
else
    github-release release \
        --user $USER \
        --repo $REPO \
        --tag $RELEASE_TAG \
        --name "$RELEASE_NAME" \
        --description "$RELEASE_DESCRIPTION"
fi

echo Uploading binaries...
for FILE in `find bin -type f`; do
    NAME=${FILE/bin\//}
    NAME=${NAME//\//-}
    NAME=`echo $NAME | sed 's/\(.*\)-local-persist/local-persist-\1/'`

    echo Uploading ${NAME}...

    if [[ $NAME != "" ]]; then
        github-release upload \
            --user $USER \
            --repo $REPO \
            --tag $RELEASE_TAG \
            --name $NAME \
            --file $FILE
    fi
done
