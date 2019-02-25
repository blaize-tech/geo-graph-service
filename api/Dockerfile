FROM golang:latest

# Copy the local package files to the container’s workspace.
ADD . /go/src/github.com/GeoServer/project/api

# Install our dependencies
RUN go get github.com/gorilla/websocket
RUN go get gopkg.in/mgo.v2
RUN go get gopkg.in/mgo.v2/bson

# Install api binary globally within container
RUN go install github.com/GeoServer/project/api

# Set binary as entrypoint
ENTRYPOINT /go/bin/api

# Expose default port (3000)
EXPOSE 3000