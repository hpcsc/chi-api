FROM golang:1.22-alpine AS build-env
WORKDIR /build

ARG VERSION=master
ADD . .
RUN GOARCH=$TARGETARCH \
    GOOS=linux \
    CGO_ENABLED=0 \
      go build --ldflags "-extldflags -static -X 'github.com/hpcsc/chi-api/internal/usecase/root.Version=${VERSION}'" \
        -o chi-api \
        ./cmd/chi-api/main.go

FROM ubuntu:22.04
WORKDIR /app

COPY --from=build-env /build/chi-api .

ENTRYPOINT ["/app/chi-api"]
