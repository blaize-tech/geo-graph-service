FROM golang:latest

# Copy the local package files to the container’s workspace.
ADD . /go/src/geo-graph-service/api

# Install our dependencies
RUN go get github.com/gorilla/websocket
RUN go get gopkg.in/mgo.v2
RUN go get gopkg.in/mgo.v2/bson

# Install api binary globally within container
RUN go install geo-graph-service/api

# Set binary as entrypoint
ENTRYPOINT /go/bin/api --port 3030

# Expose default port (3030)
EXPOSE 3030