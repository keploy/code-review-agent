# FROM golang:1.24-alpine AS builder

# WORKDIR /app
# COPY go.mod go.sum ./
# RUN go mod download
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# FROM alpine:latest
# # FIX: Add curl, which is required to download the Ollama install script
# RUN apk --no-cache add ca-certificates git bash curl

# # Configure git to trust the workspace directory
# RUN git config --global --add safe.directory /github/workspace
# RUN git config --global --add safe.directory '*'

# # Put binary in /app instead of /github/workspace
# WORKDIR /app
# COPY --from=builder /app/main .
# RUN chmod +x main

# # GitHub Actions will mount /github/workspace, but we're using /app
# ENTRYPOINT ["/app/main"]



# --- Use a standard Go image based on Debian/glibc for the builder ---
FROM golang:1.24-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build a standard Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# --- Use Ubuntu as the final stage for maximum compatibility ---
FROM ubuntu:22.04

# Install necessary packages. Ubuntu uses 'apt-get'.
# 'curl' and 'bash' are usually pre-installed, but we ensure they are present.
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    git \
    curl \
    bash \
    && rm -rf /var/lib/apt/lists/*

# Configure git to trust the workspace directory
RUN git config --global --add safe.directory /github/workspace
RUN git config --global --add safe.directory '*'

# Put binary in /app
WORKDIR /app
COPY --from=builder /app/main .
RUN chmod +x main

ENTRYPOINT ["/app/main"]