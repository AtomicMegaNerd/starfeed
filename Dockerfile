#########################################################################
# Builder image                                                         # 
#########################################################################

FROM golang:1.23.4-alpine3.21 AS builder

ENV GOTASK_VERSION=3.40.0-r0

WORKDIR /app


# for caching	copy go mod and sum files first and download dependencies
COPY ./go.mod ./go.sum /app
RUN go mod download

# copy the rest of the files
COPY . .

RUN apk add --no-cache go-task=${GOTASK_VERSION}

# Run the build
RUN go-task build

#########################################################################
# Runner image                                                          # 
#########################################################################

FROM alpine:3.21 AS runner

ENV PATH=/app/bin:$PATH
ENV USER=starfeed
ENV UID=10001
ENV GID=10001

WORKDIR /app/bin
COPY --from=builder /app/bin/starfeed /app/bin/starfeed

RUN addgroup -g $GID $USER && adduser -D -u $UID -G $USER $USER && \
    chown -R $USER:$USER /app

USER $USER

CMD ["starfeed"]

