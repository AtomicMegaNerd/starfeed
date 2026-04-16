#########################################################################
# Builder image                                                         #
#########################################################################

FROM golang:1.26.2-alpine3.23 AS builder

ARG GOTASK_VERSION="v3.50.0"

WORKDIR /app

# for caching	copy go mod and sum files first and download dependencies
COPY ./go.mod ./go.sum /app/
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download \
    && go install github.com/go-task/task/v3/cmd/task@${GOTASK_VERSION}

# copy the rest of the files
COPY . .

# Run the build
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    task build

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
ENV PATH=/app/bin:$PATH

WORKDIR /app
RUN addgroup -g ${GID} starfeed && adduser -D -u ${UID} -G starfeed starfeed
COPY --from=builder --chown=${UID}:${GID} /app/bin/starfeed /app/bin/starfeed

USER starfeed
CMD ["starfeed"]
