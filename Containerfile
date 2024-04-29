FROM quay.io/redhat-cne/openshift-origin-release:rhel-8-golang-1.20-openshift-4.14 AS builder
ENV CGO_ENABLED=1
ENV COMMON_GO_ARGS=-race
ENV GOOS=linux
ENV GOPATH=/go

WORKDIR /go/src/github.com/Jennifer-chen-rh/ptp-events-consumer
COPY . .

RUN go mod tidy
RUN go mod vendor
RUN go build

ENTRYPOINT ["./ptp-events-consumer"]
