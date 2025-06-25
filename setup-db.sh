#!/bin/bash

# M-Pesa Tinker Database Setup Script
# This script sets up the MySQL database using Docker Compose

set -e

echo "🚀 Starting M-Pesa Tinker Database Setup..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is installed and running
check_docker() {
    print_status "Checking Docker installation..."

    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! docker info &> /dev/null; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi

    print_success "Docker is installed and running"
}

# Check if Docker Compose is available
check_docker_compose() {
    print_status "Checking Docker Compose..."

    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not available. Please install Docker Compose."
        exit 1
    fi

    print_success "Docker Compose is available"
}

# Check if .env file exists
check_env_file() {
    print_status "Checking environment configuration..."

    if [ ! -f ".env" ]; then
        print_error ".env file not found. Please create it with the required database credentials."
        exit 1
    fi

    print_success "Environment file found"
}

# Start Docker Compose services
start_services() {
    print_status "Starting MySQL database..."

    # Use docker-compose or docker compose based on what's available
    if command -v docker-compose &> /dev/null; then
        docker-compose up -d
    else
        docker compose up -d
    fi

    print_success "Services started successfully"
}

# Wait for MySQL to be ready
wait_for_mysql() {
    print_status "Waiting for MySQL to be ready..."

    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if docker exec mpesa_mysql mysqladmin ping -h"localhost" -u"mpesa" -p"password" --silent; then
            print_success "MySQL is ready!"
            break
        fi

        print_status "Attempt $attempt/$max_attempts - MySQL not ready yet, waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done

    if [ $attempt -gt $max_attempts ]; then
        print_error "MySQL failed to start within expected time"
        exit 1
    fi
}

# Display connection information
show_connection_info() {
    echo ""
    echo "════════════════════════════════════════════════════════════════"
    print_success "Database setup completed successfully! 🎉"
    echo "════════════════════════════════════════════════════════════════"
    echo ""
    echo "📊 Database Connection Details:"
    echo "   Host: localhost"
    echo "   Port: 3306"
    echo "   Database: mpesa"
    echo "   Username: mpesa"
    echo "   Password: password"
    echo ""
    echo "🐳 Docker Commands:"
    echo "   Stop services:    docker-compose down"
    echo "   View logs:        docker-compose logs -f"
    echo "   Restart:          docker-compose restart"
    echo ""
    echo "🔧 Your Go application should now be able to connect to the database!"
    echo "   Make sure your .env file has the correct DB_HOST=127.0.0.1"
    echo ""
}

# Cleanup function
cleanup() {
    if [ $? -ne 0 ]; then
        print_error "Setup failed. Cleaning up..."
        if command -v docker-compose &> /dev/null; then
            docker-compose down
        else
            docker compose down
        fi
    fi
}

# Set trap for cleanup on exit
trap cleanup EXIT

# Main execution
main() {
    echo "╔════════════════════════════════════════════════════════════════╗"
    echo "║                    M-Pesa Tinker Database Setup                ║"
    echo "╚════════════════════════════════════════════════════════════════╝"
    echo ""

    check_docker
    check_docker_compose
    check_env_file
    start_services
    wait_for_mysql
    show_connection_info

    # Remove the trap since we succeeded
    trap - EXIT
}

# Run main function
main
