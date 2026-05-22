FROM golang:1.25.7 AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

ENV HTTP_ADDR=:8080
ENV HTTP_REQUEST_TIMEOUT=3s
ENV HTTP_SHUTDOWN_TIMEOUT=10s

COPY --from=builder /out/app /app

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app"]