FROM registry.ci.openshift.org/ocp/builder:rhel-8-golang-1.21-openshift-4.16 AS builder
ENV CGO_ENABLED=1
ENV COMMON_GO_ARGS=-race
ENV GOOS=linux
ENV GOPATH=/go

WORKDIR /go/src/github.com/Jennifer-chen-rh/ptp-events-consumer
COPY . .

RUN go build

ENTRYPOINT ["./ptp-events-consumer"]
