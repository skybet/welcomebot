# Build our binary
from golang:1.16 as builder
ADD . /go/src/github.com/skybet/welcomebot/
WORKDIR /go/src/github.com/skybet/welcomebot
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o welcomebot  .

