#!/bin/sh

# Load environment variables from .env file
. ./.env

# Build the docker image
docker build -t starfeed:local .

# We are going to build the command to run the docker container gradually
# based on the environment variables set in the .env file
DOCKER_CMD="docker run -e STARFEED_GITHUB_API_TOKEN=\"$STARFEED_GITHUB_API_TOKEN\" \
	-e STARFEED_FRESHRSS_URL=\"$STARFEED_FRESHRSS_URL\" \
	-e STARFEED_FRESHRSS_USER=\"$STARFEED_FRESHRSS_USER\" \
	-e STARFEED_FRESHRSS_API_TOKEN=\"$STARFEED_FRESHRSS_API_TOKEN\""

# Add the debug environment variable if it is set
if [ $STARFEED_DEBUG_MODE = "true" ]; then
	DOCKER_CMD="$DOCKER_CMD -e STARFEED_DEBUG_MODE=true"
fi

# Add single run mode if it is set
if [ $STARFEED_SINGLE_RUN_MODE = "true" ]; then
	DOCKER_CMD="$DOCKER_CMD -e STARFEED_SINGLE_RUN_MODE=true"
fi

# Add the image name
DOCKER_CMD="$DOCKER_CMD -t starfeed:local"

# Run the full command
eval "$DOCKER_CMD"

