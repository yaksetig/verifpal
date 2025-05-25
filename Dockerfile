# Use a slim Python base image
FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    git \
    pkg-config \
    libssl-dev \
  && rm -rf /var/lib/apt/lists/*

# Install Verifpal (pick correct binary for the container's CPU arch)
RUN arch="$(dpkg --print-architecture)" && \
    case "$arch" in \
      amd64) bin=verifpal_linux_amd64 ;; \
      arm64) bin=verifpal_linux_arm64 ;; \
      *) echo "Unsupported architecture: $arch" >&2; exit 1 ;; \
    esac && \
    curl -L \
      "https://github.com/symbolicsoft/verifpal/releases/download/v0.26.0/${bin}" \
      -o /usr/local/bin/verifpal && \
    chmod +x /usr/local/bin/verifpal

# Set working directory
WORKDIR /app

# Copy and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app.py .

# Expose the port your app runs on
EXPOSE 5000

# Run the application with Gunicorn
CMD ["gunicorn", "--bind", "0.0.0.0:5000", "app:app"]
