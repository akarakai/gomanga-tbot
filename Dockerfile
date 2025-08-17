FROM golang:1.24-bullseye

RUN apt-get update && apt-get install -y \
    curl \
    gnupg \
    ca-certificates \
    lsb-release \
    gcc \
    build-essential \
    libnss3 \
    libatk-bridge2.0-0 \
    libdrm2 \
    libxkbcommon0 \
    libxcomposite1 \
    libxdamage1 \
    libxrandr2 \
    libgbm1 \
    libxss1 \
    libasound2 \
    libatspi2.0-0 \
    libgtk-3-0 \
    xvfb \
    && rm -rf /var/lib/apt/lists/*

ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright
ENV PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=true

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go install github.com/playwright-community/playwright-go/cmd/playwright@v0.5200.0

RUN playwright install chromium
RUN playwright install-deps

RUN ls -la /ms-playwright/
RUN find /ms-playwright -name "*chrome*" -type f | head -5

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o main ./cmd/gomanga

CMD ["./main"]



