FROM golang:1.27-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /logflush ./cmd/logflush

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /logflush /logflush
COPY config.example.yaml /etc/logflush/config.yaml
ENTRYPOINT ["/logflush", "-config=/etc/logflush/config.yaml"]
