#############      builder       #############
FROM golang:1.14.2 AS builder

ARG TARGETS=dev

WORKDIR /go/src/github.com/onmetal/k8s-machines
COPY . .

RUN make $TARGETS

############# base
FROM alpine:3.11.3 AS base

#############      image     #############
FROM base AS image

WORKDIR /
COPY --from=builder /go/bin/machines /machines

ENTRYPOINT ["/machines"]
