FROM golang:latest

# Copy the local package files to the container’s workspace.
ADD . /go/src/github.com/GeoServer/project/api
ADD . /go/src/github.com/GeoServer/project/api/models
ADD . /go/src/github.com/GeoServer/project/api/models/item
ADD . /go/src/github.com/GeoServer/project/api/models/item/db

# Install our dependencies
RUN go get github.com/gorilla/websocket
RUN go get gopkg.in/mgo.v2
RUN go get gopkg.in/mgo.v2/bson

# Install api binary globally within container
RUN go install github.com/GeoServer/project/api

# Set binary as entrypoint
ENTRYPOINT /go/bin/api --port 3030

# Expose default port (3030)
EXPOSE 3030