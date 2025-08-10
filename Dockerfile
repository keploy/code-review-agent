FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build and verify
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
RUN ls -la main
RUN file main
RUN ./main --help || echo "Binary exists but may need runtime deps"

FROM alpine:latest
RUN apk --no-cache add ca-certificates git
RUN mkdir -p /github/workspace
COPY --from=builder /app/main /github/workspace/main
RUN chmod +x /github/workspace/main
RUN ls -la /github/workspace/main
RUN file /github/workspace/main

WORKDIR /github/workspace
ENTRYPOINT ["/github/workspace/main"]