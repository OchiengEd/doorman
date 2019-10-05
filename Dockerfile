FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git && mkdir -pv $GOPATH/src/doorman
ADD . $GOPATH/src/doorman
WORKDIR $GOPATH/src/doorman
RUN go get -d -v && go build -o /go/bin/doorman

FROM scratch
COPY --from=builder /go/bin/doorman /go/bin/doorman
ENTRYPOINT ["/go/bin/doorman"]
EXPOSE 5000
