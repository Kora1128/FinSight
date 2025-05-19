# FinSight - Personalized Investment Portfolio Tracker & Recommendation Engine

A powerful investment portfolio tracking and stock recommendation system that aggregates holdings from Zerodha and ICICI Direct, providing daily stock recommendations based on curated news feeds. The application uses PostgreSQL for persistent storage and includes user session management with email-based identification.

## Features

- **PostgreSQL Database**: Persistent storage for user data, sessions, and portfolio information
- **User Authentication**: Email-based user identification with secure session management
- **Supabase Integration**: Support for both direct PostgreSQL connection and Supabase client
- **Portfolio Aggregation**: Combine holdings from Zerodha and ICICI Direct into a unified view
- **Intelligent News Processing**: Filter financial news for relevant investment recommendations
- **Sentiment Analysis**: Analyze news articles to determine market sentiment
- **Stock Recommendations**: Get daily stock recommendations based on curated news
- **Portfolio Management**: View combined portfolio with flexible filtering options
- **In-memory Caching**: Fast data access with configurable TTL
- **RESTful API Endpoints**: Well-structured API for frontend integration

## Prerequisites

- Go 1.16+ 
- PostgreSQL database (via Supabase or standalone)
- Zerodha API Key & Secret
- ICICI Direct API Key & Secret
- OpenAI API Key (for stock symbol extraction from news)

## Environment Variables

Create a `.env` file with the following configurations:

```
# Server configuration
APP_PORT=8080
APP_ENV=development
APP_READ_TIMEOUT=10s
APP_WRITE_TIMEOUT=10s

# Zerodha API configuration
ZERODHA_API_KEY=your_zerodha_api_key
ZERODHA_API_SECRET=your_zerodha_api_secret

# ICICI Direct API configuration
ICICI_API_KEY=your_icici_api_key
ICICI_API_SECRET=your_icici_api_secret
ICICI_PASSWORD=your_icici_password

# OpenAI configuration
OPENAI_API_KEY=your_openai_api_key

# Cache configuration
CACHE_TTL=15m

# Database configuration (Supabase)
SUPABASE_URL=https://your-project-ref.supabase.co
SUPABASE_API_KEY=your_supabase_public_api_key
SUPABASE_PASSWORD=your_supabase_database_password

# News configuration
NEWS_REFRESH_INTERVAL=24h
TRUSTED_SOURCES=Economic Times,Business Standard,Moneycontrol,Livemint,Reuters India,BloombergQuint
```

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

## Running the Application

1. Create a `.env` file with your configuration (see Environment Variables section)

2. Start the server:
```bash
go run cmd/server/main.go
```

The server will start on port 8080 by default (configurable via APP_PORT environment variable).

3. Access the API using tools like cURL or Postman:
```bash
# Create a session for a user
curl -X POST -H "Content-Type: application/json" -d '{"email": "user@example.com"}' http://localhost:8080/api/v1/sessions

# Get session info (replace {userId} with an actual user ID)
curl -X GET -H "Authorization: Bearer {session_token}" http://localhost:8080/api/v1/sessions/{userId}
```

## API Endpoints

### User Session Management

- `POST /api/v1/sessions`: Create a new session with user's email
  ```json
  {
    "email": "user@example.com"
  }
  ```
- `GET /api/v1/sessions/{userId}`: Get session information for a user
- `POST /api/v1/sessions/connect`: Connect a broker to a session
  ```json
  {
    "userId": "user-id",
    "broker_type": "zerodha",
    "api_key": "your-api-key",
    "api_secret": "your-api-secret",
    "request_token": "your-request-token"
  }
  ```
- `POST /api/v1/sessions/disconnect/{userId}/{brokerType}`: Disconnect a broker from a session

### Portfolio

- `GET /api/v1/portfolio`: Retrieve aggregated portfolio data
  - Query params: `type=stock|mutualfund|all` (default: all)
- `POST /api/v1/portfolio/refresh`: Force refresh of portfolio data

### Recommendations

- `GET /api/v1/recommendations`: Get all stock recommendations
- `GET /api/v1/recommendations/latest`: Get latest recommendations
- `GET /api/v1/recommendations/stock/:symbol`: Get recommendations for a specific stock

### News Sources

- `GET /api/v1/news/sources`: Get all configured news sources
- `POST /api/v1/news/sources`: Add a new news source
- `DELETE /api/v1/news/sources/:name`: Remove a news source

## Project Structure

```
FinSight/
├── cmd/
│   └── server/           # Application entry point
│       └── main.go
├── internal/
│   ├── api/              # API handlers, middleware, and routes
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── routes/
│   ├── broker/           # Broker integrations
│   │   ├── icici_direct/ # ICICI Direct API integration
│   │   └── zerodha/      # Zerodha API integration
│   ├── cache/            # Cache implementation
│   ├── config/           # Application configuration
│   ├── models/           # Data models
│   ├── news/             # News processing and recommendation engine
│   └── portfolio/        # Portfolio aggregation service
└── pkg/                  # Shared packages
    ├── logger/           # Logging utilities
    └── utils/            # General utilities
```

## Dependencies

- [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin): Web framework
- [github.com/zerodha/gokiteconnect](https://github.com/zerodha/gokiteconnect): Official Zerodha API client
- [github.com/Kora1128/icici-breezeconnect-go](https://github.com/Kora1128/icici-breezeconnect-go): Custom ICICI Direct API client
- [github.com/mmcdole/gofeed](https://github.com/mmcdole/gofeed): RSS feed parser
- [github.com/patrickmn/go-cache](https://github.com/patrickmn/go-cache): In-memory caching
- [github.com/sashabaranov/go-openai](https://github.com/sashabaranov/go-openai): OpenAI API client
- [github.com/lib/pq](https://github.com/lib/pq): PostgreSQL driver
- [github.com/google/uuid](https://github.com/google/uuid): UUID generation for sessions and users
- [github.com/joho/godotenv](https://github.com/joho/godotenv): Loading environment variables from .env files
- [github.com/supabase-community/supabase-go](https://github.com/supabase-community/supabase-go): Supabase Go client

## Future Enhancements

- Enhanced user authentication with password support
- Implementation of broker reconnection workflows
- Data visualization for portfolio performance
- Mobile app integration
- Custom watchlists and alerts
- Advanced technical analysis
- Historical portfolio tracking and performance analysis

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

- Broker API credentials are stored securely in the database
- All API endpoints are protected with session-based authentication
- Email is used for user identification, plan to add password authentication in future
- HTTPS is recommended for production deployment
- Environment variables should be kept secure and not committed to version control

## Recent Updates

- **Database Migration**: Migrated from SQLite to PostgreSQL for improved scalability and performance
- **User Management**: Added email-based user identification and session management
- **Supabase Integration**: Added support for Supabase client and direct PostgreSQL connections
- **Environment Configuration**: Enhanced environment configuration with Supabase support
- **Session Persistence**: Sessions and broker credentials are now stored persistently

## Current Limitations

- Limited historical data tracking
- Basic email-only authentication (no passwords yet)
- No refresh token mechanism for broker connections

## License

MIT License - see LICENSE file for details