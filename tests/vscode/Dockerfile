FROM golang:1.25-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /temporal-lsp ./cmd/temporal-lsp/

FROM codercom/code-server:latest

USER root

# Install system deps for Playwright Chromium + ffmpeg for GIF conversion
RUN apt-get update && apt-get install -y \
    curl ffmpeg \
    libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 \
    libdrm2 libxkbcommon0 libxcomposite1 libxdamage1 \
    libxfixes3 libxrandr2 libgbm1 libpango-1.0-0 \
    libcairo2 libasound2 libatspi2.0-0 libxshmfence1 \
    && rm -rf /var/lib/apt/lists/*

# Install Node 20 (code-server ships an older one)
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install temporal-lsp binary
COPY --from=builder /temporal-lsp /usr/local/bin/temporal-lsp

# Install the VS Code extension
COPY integrations/vscode/temporal-lsp-0.1.0.vsix /tmp/temporal-lsp.vsix
RUN code-server --install-extension /tmp/temporal-lsp.vsix

# Configure code-server settings for the LSP
RUN mkdir -p /root/.local/share/code-server/User
COPY tests/vscode/settings.json /root/.local/share/code-server/User/settings.json

# Copy focused demo workspace files (one violation per file, all flat)
COPY tests/vscode/workspace/ /workspace/

# Set up Playwright
WORKDIR /playwright
COPY tests/vscode/package.json ./
RUN npm install
RUN npx playwright install chromium --with-deps

COPY tests/vscode/playwright.config.ts tests/vscode/screenshot.spec.ts ./

# Entrypoint: start code-server, wait, run Playwright
COPY tests/vscode/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
