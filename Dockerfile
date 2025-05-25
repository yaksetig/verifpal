FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Download and install Verifpal
# First, let's check the architecture and download the correct binary
RUN ARCH=$(uname -m) && \
    echo "Architecture: $ARCH" && \
    if [ "$ARCH" = "x86_64" ]; then \
        wget -O /usr/local/bin/verifpal https://github.com/symbolicsoft/verifpal/releases/download/v0.26.0/verifpal_linux_amd64; \
    elif [ "$ARCH" = "aarch64" ]; then \
        wget -O /usr/local/bin/verifpal https://github.com/symbolicsoft/verifpal/releases/download/v0.26.0/verifpal_linux_arm64; \
    else \
        echo "Unsupported architecture: $ARCH" && exit 1; \
    fi && \
    chmod +x /usr/local/bin/verifpal && \
    # Test if it works
    /usr/local/bin/verifpal help || echo "Verifpal test failed, continuing anyway"

# Alternative approach: Build from source if the binary doesn't work
# Uncomment the following lines if the above doesn't work:
# RUN apt-get update && apt-get install -y golang-go git && \
#     git clone https://github.com/symbolicsoft/verifpal.git && \
#     cd verifpal && \
#     go build -o /usr/local/bin/verifpal cmd/verifpal/main.go && \
#     chmod +x /usr/local/bin/verifpal && \
#     cd .. && rm -rf verifpal

# Set working directory
WORKDIR /app

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app.py .

# Expose port
EXPOSE 5000

# Debug: Check verifpal installation
RUN which verifpal && verifpal help || echo "Verifpal not found in PATH"

# Run the application
CMD ["gunicorn", "--bind", "0.0.0.0:5000", "app:app"]
