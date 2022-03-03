# To build this image run on this directory
#   docker build ./../.. -f btcd-dockertests.Dockerfile -t btcd-dockertests
# To run this image without network address
#   docker run --net none -it btcd-dockertests

#### Build btcd
FROM golang:1.16-alpine AS build-container


ADD . /app/src
WORKDIR /app/src
RUN mkdir /app/bin
RUN go build -o /app/bin/btcd

#### Build btcd docker image
FROM alpine:3.12

COPY --from=build-container /app/bin/btcd /app/bin/btcd

# 8333  Mainnet Bitcoin peer-to-peer port
# 8334  Mainet RPC port
EXPOSE 8333 8334

ENTRYPOINT ["/app/bin/btcd"]
