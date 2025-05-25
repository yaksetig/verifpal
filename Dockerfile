FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    git \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Verifpal
RUN curl -L https://github.com/symbolicsoft/verifpal/releases/download/v0.26.0/verifpal_linux_amd64 \
     -o /usr/local/bin/verifpal \
    && chmod +x /usr/local/bin/verifpal

# Set working directory
WORKDIR /app

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app.py .

# Expose port
EXPOSE 5000

# Run the application
CMD ["gunicorn", "--bind", "0.0.0.0:5000", "app:app"]
