#!/usr/bin/env bash

set -e

VERSION="v1.0.0-beta.2"

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
                ARCH="386"
                SUPPORTED=true
            ;;
            # untested
            arm*)
                ARCH="arm"
                SUPPORTED=true
            ;;
        esac
    elif [[ $OS == 'freebsd' ]]; then
        # untested for 32 bit
        ARCH=$(uname -m)
        SUPPORTED=true
    fi

    if [[ $SUPPORTED == false ]]; then
        echo $OS $ARCH is not supported
        exit 2
    fi
}

function install-binary {
    echo Stopping docker-volume-local-persist service if running
    if [[ $* == *--upstart* ]]; then
        sudo service docker-volume-local-persist stop
    else
        sudo systemctl stop docker-volume-local-persist
    fi

    BINARY_URL="https://github.com/CWSpear/local-persist/releases/download/${VERSION}/local-persist-${OS}-${ARCH}"

    echo Downloading binary:
    echo "  From: $BINARY_URL"
    echo "  To:   /usr/local/bin/docker-volume-local-persist"

    curl -fLsS "$BINARY_URL" > /usr/local/bin/docker-volume-local-persist
    chmod +x /usr/local/bin/docker-volume-local-persist

    echo Binary download
}

function setup-upstart {
    UPSTART_CONFIG_URL="https://raw.githubusercontent.com/CWSpear/local-persist/${VERSION}/init/upstart.conf"

    echo Downloading binary:
    echo "  From: $UPSTART_CONFIG_URL"
    echo "  To:   /etc/init/docker-volume-local-persist.conf"

    sudo curl -fLsS "$UPSTART_CONFIG_URL" > /etc/init/docker-volume-local-persist.conf

    echo Upstart conf downloaded
}

function start-upstart {
    echo Reloading Upstart config and starting docker-volume-local-persist service...

    sudo initctl reload-configuration
    sudo service docker-volume-local-persist start
    sudo service docker-volume-local-persist status

    echo Done
}

function setup-systemd {
    SYSTEMD_CONFIG_URL="https://raw.githubusercontent.com/CWSpear/local-persist/${VERSION}/init/systemd.service"

    echo Downloading Systemd service conf:
    echo "  From: ${SYSTEMD_CONFIG_URL}"
    echo "  To:   /etc/systemd/system/docker-volume-local-persist.service"

    sudo curl -fLsS "$SYSTEMD_CONFIG_URL" > /etc/systemd/system/docker-volume-local-persist.service

    echo Systemd conf downloaded
}

function start-systemd {
    echo Starting docker-volume-local-persist service...

    sudo systemctl daemon-reload
    sudo systemctl enable docker-volume-local-persist
    sudo systemctl start docker-volume-local-persist
    sudo systemctl status docker-volume-local-persist

    echo Done
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
