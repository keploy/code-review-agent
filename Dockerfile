FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates git

# Configure git to trust the workspace directory
RUN git config --global --add safe.directory /github/workspace
RUN git config --global --add safe.directory '*'

# Put binary in /app instead of /github/workspace
WORKDIR /app
COPY --from=builder /app/main .
RUN chmod +x main

# GitHub Actions will mount /github/workspace, but we're using /app
ENTRYPOINT ["/app/main"]