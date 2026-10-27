FROM --platform=$BUILDPLATFORM golang:1.27.3-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /logflush ./cmd/logflush

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /logflush /logflush

USER nonroot:nonroot

ENTRYPOINT ["/logflush", "-config=/etc/logflush/config.yaml"]
