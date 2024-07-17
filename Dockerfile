FROM golang:1.22.5-alpine3.20 as builder

ARG GOLANG_NAMESPACE="github.com/lxbot/lxbot"
ENV GOLANG_NAMESPACE="$GOLANG_NAMESPACE"

RUN apk --no-cache add alpine-sdk coreutils make tzdata
RUN cp -f /usr/share/zoneinfo/Asia/Tokyo /etc/localtime
WORKDIR /go/src/$GOLANG_NAMESPACE
ADD ./go.* /go/src/$GOLANG_NAMESPACE/
ENV GO111MODULE=on
RUN go mod download
ADD . /go/src/$GOLANG_NAMESPACE/
RUN make build
RUN mkdir -p /lxbot
RUN mv /go/src/$GOLANG_NAMESPACE/lxbot /lxbot/

# ====================================================================================

FROM alpine:3.20

RUN apk --no-cache add ca-certificates docker
COPY --from=builder /etc/localtime /etc/localtime
COPY --from=builder /lxbot /lxbot

WORKDIR /lxbot
CMD ["/lxbot/lxbot"]
