FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-1.22-openshift-4.17 AS builder
ENV CGO_ENABLED=1
ENV COMMON_GO_ARGS=-race
ENV GOOS=linux
ENV GOPATH=/go

WORKDIR /go/src/github.com/Jennifer-chen-rh/ptp-events-consumer
COPY . .

RUN go build

FROM --platform=linux/x86_64 registry.ci.openshift.org/ocp/4.17:base-rhel9 AS bin
COPY --from=builder /go/src/github.com/Jennifer-chen-rh/ptp-events-consumer/ptp-events-consumer /

ENTRYPOINT ["./ptp-events-consumer"]
