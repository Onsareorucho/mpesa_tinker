# M-Pesa Tinker API

A comprehensive Go-based REST API for testing M-Pesa integrations, featuring STK Push payments, URL registration, QR code generation, and transaction status tracking with MySQL database storage.

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for development)
- MySQL (via Docker)

### 1. Setup Environment

Create a `.env` file in the project root:

```env
# Database Configuration
DB_HOST=127.0.0.1
DB_PORT=3305
DB_USER=mpesa
DB_PASS=password
DB_NAME=mpesa

# M-Pesa Sandbox Configuration
MPESA_CONSUMER_KEY=your_consumer_key
MPESA_CONSUMER_SECRET=your_consumer_secret
MPESA_SHORTCODE=174379
MPESA_PASSKEY=your_passkey
MPESA_CALLBACK_URL=https://yourdomain.com/callback

# Server Configuration
PORT=8080
```

## Start Database

```sh
# Make the setup script executable and run it
chmod +x setup-db.sh
./setup-db.sh

```

Run the api server

```sh
# Install dependencies
go mod tidy

# Run the server
go run main.go

```

The server should now be running on `http://localhost:8080`.

API Endpoints

1. STK push payment
   Initiate an mpesa STK push payment request

Endpoint `/stk-push`

```json
{
  "phone": "254712345678",
  "amount": "100"
}
```

exmaple using curl

```sh
curl -X POST http://localhost:8080/stkpush \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "254712345678",
    "amount": "100"
  }'
```
