FROM golang:1.22.4
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 go install -a -ldflags '-extldflags "-static"' ./cmd/k8s-heartbeat

FROM alpine
COPY --from=0 /go/bin/k8s-heartbeat /usr/bin/
EXPOSE 8080
CMD ["k8s-heartbeat"]

