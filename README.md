# FinSight - Investment Portfolio Tracker & Recommendation Engine

A powerful investment portfolio tracking and stock recommendation system that aggregates holdings from Zerodha and ICICI Direct, providing daily stock recommendations based on curated news feeds.

## Features

- Portfolio aggregation from Zerodha and ICICI Direct
- Real-time portfolio tracking
- Daily stock recommendations from trusted sources
- In-memory caching for fast data access
- RESTful API endpoints for frontend integration

## Prerequisites

- Go 1.21 or later
- Zerodha API credentials (API Key and Secret)
- ICICI Direct API credentials (API Key, Secret, and Password)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/Kora1128/FinSight.git
cd FinSight
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
# Zerodha API Credentials
export ZERODHA_API_KEY="your_api_key"
export ZERODHA_API_SECRET="your_api_secret"

# ICICI Direct API Credentials
export ICICI_API_KEY="your_api_key"
export ICICI_API_SECRET="your_api_secret"
export ICICI_PASSWORD="your_password"

# Application Configuration
export APP_PORT="8080"
export APP_ENV="development" # or "production"
```

## Running the Application

1. Start the server:
```bash
go run cmd/server/main.go
```

The server will start on port 8080 by default (configurable via APP_PORT environment variable).

## API Endpoints

### Authentication
- `POST /api/v1/login/zerodha` - Zerodha login
- `POST /api/v1/login/icici` - ICICI Direct login

### Portfolio
- `GET /api/v1/portfolio` - Get aggregated portfolio
  - Query Parameters:
    - `type`: stock|mutualfund|all (default: all)
- `POST /api/v1/portfolio/refresh` - Refresh portfolio data

### Recommendations
- `GET /api/v1/recommendations` - Get daily stock recommendations

### User Status
- `GET /api/v1/user/status` - Check login status for brokers

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── routes/
│   ├── broker/
│   │   ├── zerodha/
│   │   └── icici/
│   ├── cache/
│   ├── config/
│   ├── models/
│   └── news/
├── pkg/
│   ├── logger/
│   └── utils/
├── go.mod
├── go.sum
└── README.md
```

## Development

### Adding New Features

1. Create feature branch:
```bash
git checkout -b feature/your-feature-name
```

2. Make changes and commit:
```bash
git add .
git commit -m "feat: your feature description"
```

3. Push changes:
```bash
git push origin feature/your-feature-name
```

### Running Tests

```bash
go test ./...
```

## Security Notes

- API credentials are stored in memory only during the session
- All API endpoints are protected with appropriate authentication
- HTTPS is recommended for production deployment

## Limitations (Phase 1)

- Data is stored in-memory and will be lost on application restart
- No persistent storage
- Limited historical data tracking

## License

MIT License - see LICENSE file for details