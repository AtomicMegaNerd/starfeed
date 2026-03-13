#########################################################################
# Builder image                                                         #
#########################################################################

FROM golang:1.26.1-alpine3.23 AS builder

ENV GOTASK_VERSION=3.45.5-r4
ENV CGO_ENABLED=0
ENV GOFLAGS=-ldflags=-s\ -w

WORKDIR /app

# for caching	copy go mod and sum files first and download dependencies
COPY ./go.mod ./go.sum /app/
RUN go mod download

# copy the rest of the files
COPY . .

RUN apk add --no-cache go-task=${GOTASK_VERSION}

# Run the build
RUN go-task build

#########################################################################
# Runner image                                                          #
#########################################################################

FROM alpine:3.23 AS runner

LABEL org.opencontainers.image.title="starfeed"
LABEL org.opencontainers.image.description="Starfeed subsribes to RSS feeds for starred GitHub repos"
LABEL org.opencontainers.image.authors="Chris Dunphy"
LABEL org.opencontainers.image.source="https://github.com/atomicmeganerd/starfeed"
LABEL org.opencontainers.image.licenses="MIT"

ARG UID=10001
ARG GID=10001
ENV USER=starfeed
ENV UID=${UID}
ENV GID=${GID}

WORKDIR /app
ENV PATH=/app/bin:$PATH
COPY --from=builder --chown=${UID}:${GID} /app/bin/starfeed /app/bin/starfeed

RUN addgroup -g $GID $USER && adduser -D -u $UID -G $USER $USER

USER $USER

CMD ["starfeed"]
