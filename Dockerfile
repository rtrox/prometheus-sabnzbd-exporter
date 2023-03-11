ARG VERSION="devel"

FROM golang:1.20-alpine AS build_base
WORKDIR /tmp/sabnzbd-exporter

COPY . .

RUN go mod tidy && \
    go mod vendor && \
    CGO_ENABLED=0 go build \
        -ldflags="-s -w -X \"main.version=${VERSION}\"" \
        -o ./out/sabnzbd-exporter \
         cmd/sabnzbd-exporter/sabnzbd-exporter.go

FROM scratch
COPY --from=build_base /tmp/sabnzbd-exporter/out/sabnzbd-exporter /bin/sabnzbd-exporter
EXPOSE 8080/tcp
ENTRYPOINT ["/bin/sabnzbd-exporter"]