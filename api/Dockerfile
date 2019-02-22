FROM golang:latest

# Copy the local package files to the container’s workspace.
ADD ./api /api

# Install our dependencies
RUN go get github.com/gorilla/websocket
RUN go get gopkg.in/mgo.v2
RUN go get gopkg.in/mgo.v2/bson

# Set binary as entrypoint
#ENTRYPOINT /go/bin/api

# Expose default port (3000)
EXPOSE 3000