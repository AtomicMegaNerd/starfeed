#!/bin/sh

. ./.env

docker build -t starfeed:local .

docker run -e GITHUB_API_TOKEN=${GITHUB_API_TOKEN} \
	-e FRESHRSS_URL=${FRESHRSS_URL} \
	-e FRESHRSS_USER=${FRESHRSS_USER} \
	-e FRESHRSS_API_TOKEN=${FRESHRSS_API_TOKEN} \
	-e STARFEED_DEBUG=${STARFEED_DEBUG} \
	-t starfeed:local


