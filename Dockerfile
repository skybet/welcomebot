# Build our binary
from golang:1.10 as builder
RUN go get github.com/skybet/welcomebot
WORKDIR /go/src/github.com/skybet/welcomebot
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o welcomebot  .

# Build our final image
FROM alpine:latest
RUN apk add -U ca-certificates
COPY --from=builder /go/src/github.com/skybet/welcomebot/welcomebot /welcomebot
CMD ["/welcomebot"]
