FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git && mkdir -pv $GOPATH/src/doorman
ADD . $GOPATH/src/doorman
WORKDIR $GOPATH/src/doorman
RUN go get -d -v && go build -o /go/bin/doorman

FROM alpine:latest
WORKDIR /app
COPY --from=builder /go/bin/doorman /app/doorman
ENTRYPOINT ["./doorman"]
EXPOSE 5000
