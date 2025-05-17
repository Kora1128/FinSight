# FinSight - Personalized Investment Portfolio Tracker & Recommendation Engine

A powerful investment portfolio tracking and stock recommendation system that aggregates holdings from Zerodha and ICICI Direct, providing daily stock recommendations based on curated news feeds.

## Features

- **Portfolio Aggregation**: Combine holdings from Zerodha and ICICI Direct into a unified view
- **Intelligent News Processing**: Filter financial news for relevant investment recommendations
- **Sentiment Analysis**: Analyze news articles to determine market sentiment
- **Stock Recommendations**: Get daily stock recommendations based on curated news
- **Portfolio Management**: View combined portfolio with flexible filtering options
- **In-memory Caching**: Fast data access with configurable TTL
- **RESTful API Endpoints**: Well-structured API for frontend integration

## Prerequisites

- Go 1.16+ 
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

1. Start the server:
```bash
go run cmd/server/main.go
```

The server will start on port 8080 by default (configurable via APP_PORT environment variable).

## API Endpoints

### Authentication

- `POST /api/v1/login/zerodha`: Authenticate with Zerodha
- `POST /api/v1/login/icici`: Authenticate with ICICI Direct
- `GET /api/v1/user/status`: Check login status for both brokers

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

## Future Enhancements

- Database integration for persistent storage
- User authentication and multi-user support
- Data visualization for portfolio performance
- Mobile app integration
- Custom watchlists and alerts
- Advanced technical analysis

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