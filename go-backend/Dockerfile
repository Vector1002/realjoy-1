# Use the official Golang base image
FROM golang:1.21

# Install Chromium and required dependencies
RUN apt-get update && apt-get install -y \
    chromium \
    wget \
    unzip \
    curl \
    gnupg \
    ca-certificates \
    fonts-liberation \
    libappindicator3-1 \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libcups2 \
    libdbus-1-3 \
    libgdk-pixbuf2.0-0 \
    libnspr4 \
    libnss3 \
    libx11-xcb1 \
    libxcomposite1 \
    libxdamage1 \
    libxrandr2 \
    xdg-utils \
    libu2f-udev \
    libvulkan1 \
    libxss1 \
    --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

# Set environment variable for Rod to find Chrome
ENV ROD_BROWSER_PATH=/usr/bin/chromium

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Build the Go app
RUN go build -o app .

# Port to expose (you can also use ENV PORT)
EXPOSE 10000

# Run the binary
CMD ["./app"]
