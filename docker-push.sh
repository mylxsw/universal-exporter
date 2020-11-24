#!/usr/bin/env bash

TAG=`cat VERSION`

docker build -t mylxsw/universal-exporter .

docker tag mylxsw/universal-exporter mylxsw/universal-exporter:$TAG
docker tag mylxsw/universal-exporter:$TAG mylxsw/universal-exporter:latest
docker push mylxsw/universal-exporter:$TAG
docker push mylxsw/universal-exporter:latest

