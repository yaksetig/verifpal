FROM golang:1.21-slim AS builder

# Install git
RUN apt-get update && apt-get install -y git

# Clone and build Verifpal
RUN git clone https://github.com/symbolicsoft/verifpal.git /verifpal
WORKDIR /verifpal
RUN go build -o verifpal cmd/verifpal/main.go

# Final stage
FROM python:3.11-slim

# Copy Verifpal binary from builder
COPY --from=builder /verifpal/verifpal /usr/local/bin/verifpal
RUN chmod +x /usr/local/bin/verifpal

# Set working directory
WORKDIR /app

# Copy and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app.py .

# Expose port
EXPOSE 5000

# Run the application
CMD ["gunicorn", "--bind", "0.0.0.0:5000", "app:app"]
