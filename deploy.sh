#!/bin/sh

NAME=web-api
BASE_DIR=$(pwd)
RELEASE=${BASE_DIR}/release
TARGET_DEPLOY=${RELEASE}/deploy
imageID=$1
mkdir -p ${TARGET_DEPLOY}
cp -rf ${BASE_DIR}/build/package/deploy.yaml ${TARGET_DEPLOY}
sed -i "s#{{.Image}}#${imageID}#g" ${TARGET_DEPLOY}/deploy.yaml