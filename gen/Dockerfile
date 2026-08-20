ARG GO_VERSION=1.26

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS runner
ARG TARGETOS TARGETARCH
WORKDIR /r

COPY go.mod go.sum ./
COPY ./gen/main.go ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags='-s -w' -o /runner main.go

FROM scratch
COPY --from=runner /runner /runner
ENTRYPOINT ["/runner"]
