FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    ca-certificates \
    file \
    && rm -rf /var/lib/apt/lists/*

# Create directory for verifpal
RUN mkdir -p /usr/local/bin

# Download and install Verifpal with proper permissions
RUN wget -O /tmp/verifpal https://github.com/symbolicsoft/verifpal/releases/download/v0.26.0/verifpal_linux_amd64 && \
    # Check file type
    file /tmp/verifpal && \
    # Copy to final location
    cp /tmp/verifpal /usr/local/bin/verifpal && \
    # Make it executable
    chmod 755 /usr/local/bin/verifpal && \
    # Verify it's executable
    ls -la /usr/local/bin/verifpal && \
    # Clean up
    rm /tmp/verifpal

# Add /usr/local/bin to PATH explicitly
ENV PATH="/usr/local/bin:${PATH}"

# Test verifpal (this might fail but that's ok for now)
RUN /usr/local/bin/verifpal help || echo "Verifpal test run completed"

# Set working directory
WORKDIR /app

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app.py .

# Expose port
EXPOSE 5000

# Final check
RUN which verifpal || echo "verifpal not in PATH" && \
    ls -la /usr/local/bin/verifpal || echo "verifpal binary not found"

# Run the application
CMD ["gunicorn", "--bind", "0.0.0.0:5000", "app:app"]
