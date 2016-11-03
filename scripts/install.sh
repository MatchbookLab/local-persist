#!/usr/bin/env bash

set -e

VERSION="v1.3.0"

# uname -s, uname -m
# Deb 32: Linux i686
# Ubuntu 64: Linux x86_64
# FreeBSD: FreeBSD amd64

if [[ "$UID" != 0 ]]; then
    echo NOTE: sudo needed to set up and run start service
    exit 1
fi


function setenv {
    OS=$(uname -s | tr "[:upper:]" "[:lower:]")
    ARCH=$(uname -m)

    SUPPORTED=false
    if [[ $OS == "linux" ]]; then
        case $ARCH in
            "x86_64")
                ARCH="amd64"
                SUPPORTED=true
            ;;
            "i686")
                # ARCH="386"
                SUPPORTED=false
            ;;
            # untested
            arm*)
                # ARCH="arm"
                SUPPORTED=false
            ;;
        esac
    elif [[ $OS == 'freebsd' ]]; then
        ARCH=$(uname -m)
        SUPPORTED=false
    fi

    if [[ $SUPPORTED == false ]]; then
        echo $OS $ARCH is not supported
        exit 2
    fi
}

function install-binary {
    echo Stopping docker-volume-local-persist service if running
    echo ''
    if [[ $* == *--upstart* ]]; then
        (sudo service docker-volume-local-persist stop || true)
    else
        (sudo systemctl stop docker-volume-local-persist || true)
    fi

    BINARY_URL="https://github.com/CWSpear/local-persist/releases/download/${VERSION}/local-persist-${OS}-${ARCH}"
    BINARY_DEST="/usr/bin/docker-volume-local-persist"

    echo Downloading binary:
    echo "  From: $BINARY_URL"
    echo "  To:   $BINARY_DEST"

    curl -fLsS "$BINARY_URL" > $BINARY_DEST
    chmod +x $BINARY_DEST

    echo Binary download
    echo ''
}

# Systemd (default)
function setup-systemd {
    SYSTEMD_CONFIG_URL="https://raw.githubusercontent.com/CWSpear/local-persist/${VERSION}/init/systemd.service"
    SYSTEMD_CONFIG_DEST="/etc/systemd/system/docker-volume-local-persist.service"

    echo Downloading Systemd service conf:
    echo "  From: $SYSTEMD_CONFIG_URL"
    echo "  To:   $SYSTEMD_CONFIG_DEST"

    sudo curl -fLsS "$SYSTEMD_CONFIG_URL" > $SYSTEMD_CONFIG_DEST

    echo Systemd conf downloaded
    echo ''
}

function start-systemd {
    echo Starting docker-volume-local-persist service...

    sudo systemctl daemon-reload
    sudo systemctl enable docker-volume-local-persist
    sudo systemctl start docker-volume-local-persist
    sudo systemctl status --full --no-pager docker-volume-local-persist

    echo ''
    echo Done! If you see this message, that should mean everything is installed and is running.
}

# Upstart
function setup-upstart {
    UPSTART_CONFIG_URL="https://raw.githubusercontent.com/CWSpear/local-persist/${VERSION}/init/upstart.conf"
    UPSTART_CONFIG_DEST="/etc/init/docker-volume-local-persist.conf"

    echo Downloading binary:
    echo "  From: $UPSTART_CONFIG_URL"
    echo "  To:   $UPSTART_CONFIG_DEST"

    sudo curl -fLsS "$UPSTART_CONFIG_URL" > $UPSTART_CONFIG_DEST

    echo Upstart conf downloaded
    echo ''
}

function start-upstart {
    echo Reloading Upstart config and starting docker-volume-local-persist service...

    sudo initctl reload-configuration
    sudo service docker-volume-local-persist start
    sudo service docker-volume-local-persist status

    echo ''
    echo Done! If you see this message, that should mean everything is installed and is running.
}


setenv

if [[ $* == *--upstart* ]]; then
    install-binary --upstart
    setup-upstart
    start-upstart
else
    install-binary
    setup-systemd
    start-systemd
fi
