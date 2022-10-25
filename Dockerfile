FROM golang:latest AS build-env
WORKDIR /build/src
ENV GOPATH /build
ENV CGO_ENABLED 0
RUN apt-get update && apt-get install -y tzdata

RUN curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s latest
RUN curl https://github.com/sonatype-nexus-community/nancy/releases/download/v1.0.41/nancy-v1.0.41-linux-amd64 -o bin/nancy && chmod +x bin/nancy

# optimize dependency download
ADD src/go.* /build/src/
RUN go mod download

ADD . /build/
RUN go get
RUN go build -a -ldflags="-X main.version=$(git describe --always --long --tags)"
RUN mkdir coverage; go test -p 1 -v -covermode=atomic -coverprofile=coverage/coverage.out ./...
RUN bin/gosec .
RUN go mod vendor -v 2>&1 | grep -Eoh "([a-z0-9]|\/|\.|\_|\-)+ v.+$" | bin/nancy sleuth && rm -rf vendor

FROM scratch
COPY --from=build-env /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build-env /build/src/thermia /
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /build/src/go.mod /
ENV TZ="Europe/Helsinki"
ENTRYPOINT ["./thermia"]