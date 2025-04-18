# Start from an official Go image
FROM golang:1.21

# Install Chromium for Rod
RUN apt-get update && apt-get install -y chromium

# Debug: Check if chromium is installed
RUN which chromium && chromium --version

# Set environment variables for Rod to use Chromium
ENV CHROME_BIN=/usr/bin/chromium
ENV ROD_BROWSER_PATH=/usr/bin/chromium

WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download
# Copy everything to container
COPY . .
# Build the Go app
RUN go build -o go-backend

# Expose port (used by Render)
EXPOSE 8080

# Run the app
CMD ["./go-backend"]

