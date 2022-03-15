# GitHub action dockerfile
# Requires docker experimental features as buildx and BuildKit
# buildx cli plugin not enabled by default on ubuntu
# https://docs.docker.com/develop/develop-images/build_enhancements/#to-enable-buildkit-builds

###########################
# Build binaries stage
###########################
FROM --platform=$BUILDPLATFORM golang@sha256:db2475a1dbb2149508e5db31d7d77a75e6600d54be645f37681f03f2762169ba AS build
WORKDIR /app
# Arguments required to build binaries targetting the correct OS and CPU architectures
ARG TARGETOS TARGETARCH
# Actually building the binaries using caching
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go install -v . ./cmd/...

###########################
# Build final image stage
###########################
FROM alpine
COPY --from=build /go/bin /bin
# 8333  Mainnet Bitcoin peer-to-peer port
# 8334  Mainet RPC port
EXPOSE 8333 8334
ENTRYPOINT ["btcd"]