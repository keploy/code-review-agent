FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates git

# Create the directory and copy the binary
RUN mkdir -p /github/workspace
COPY --from=builder /app/main /github/workspace/main
RUN chmod +x /github/workspace/main

# Set working directory last
WORKDIR /github/workspace

# Use absolute path to be sure
ENTRYPOINT ["/github/workspace/main"]