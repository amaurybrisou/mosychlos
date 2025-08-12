# FinRobot API Server Dockerfile
FROM python:3.12-slim

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements and install Python dependencies
COPY requirements-minimal.txt ./
RUN pip install --no-cache-dir -r requirements-minimal.txt

# Install older pyautogen version for compatibility
RUN pip install "pyautogen==0.2.32"

# Copy FinRobot application files
COPY . ./

# Create necessary directories
RUN mkdir -p data reports

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Start the API server
CMD ["python", "api_server.py"]
