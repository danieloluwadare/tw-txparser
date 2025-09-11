# TW Transaction Parser

A high-performance Ethereum transaction parser that continuously monitors blockchain blocks, extracts transaction data, and provides a REST API for querying transactions by address. The system features intelligent backward scanning, real-time forward polling, and in-memory storage for fast data access.

## 🏗️ Architecture Overview

The system consists of four main components:

- **Parser/Poller**: Core engine that monitors blockchain and processes transactions
- **RPC Client**: Communicates with Ethereum nodes via JSON-RPC
- **Storage**: In-memory data store for subscriptions and transactions
- **HTTP Server**: REST API for external access

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Server   │    │  Parser/Poller  │    │   RPC Client    │
│                 │    │                 │    │                 │
│  - /subscribe   │◄───┤  - Backward     │◄───┤  - eth_blockNumber│
│  - /current     │    │    Scanning     │    │  - eth_getBlock │
│  - /transactions│    │  - Forward      │    │                 │
│                 │    │    Polling      │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │
         │                       ▼
         │              ┌─────────────────┐
         └──────────────►│  Memory Storage │
                        │                 │
                        │  - Subscriptions│
                        │  - Transactions │
                        └─────────────────┘
```

## 🚀 Features

- **Real-time Block Monitoring**: Continuously polls for new blocks
- **Historical Data Scanning**: Configurable backward scanning for missed blocks
- **Address Subscription**: Track specific Ethereum addresses
- **Transaction Indexing**: Stores both incoming and outgoing transactions per address
- **REST API**: Simple HTTP endpoints for data access
- **Graceful Shutdown**: Handles SIGINT/SIGTERM signals properly
- **Configurable Behavior**: Environment-based configuration
- **Docker Support**: Multi-stage build with optimized production image
- **Health Checks**: Built-in health monitoring for container orchestration

## 📋 Prerequisites

- Go 1.24 or later (for native installation)
- Docker and Docker Compose (for containerized deployment)
- Access to an Ethereum RPC endpoint

## 🛠️ Installation

### Option 1: Docker (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/danieloluwadare/tw-txparser.git
cd tw-txparser
```

2. Build and run with Docker:
```bash
# Build the image
docker build -t tw-txparser .

# Run the container
docker run -p 8080:8080 \
  -e BACKWARD_SCAN_ENABLED=true \
  -e BACKWARD_SCAN_DEPTH=10000 \
  tw-txparser
```

3. Or use Docker Compose:
```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

### Option 2: Native Go Installation

1. Clone the repository:
```bash
git clone https://github.com/danieloluwadare/tw-txparser.git
cd tw-txparser
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o txparser ./cmd/txparser
```

## ⚙️ Configuration

The parser supports the following environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `BACKWARD_SCAN_ENABLED` | `true` | Enable/disable historical block scanning |
| `BACKWARD_SCAN_DEPTH` | `10000` | Number of blocks to scan backward from current |

### Example Configuration

```bash
export BACKWARD_SCAN_ENABLED=true
export BACKWARD_SCAN_DEPTH=5000
./txparser
```

## 🏃‍♂️ Running the Application

### Docker
```bash
# Using Docker Compose (recommended)
docker-compose up -d

# Or using Docker directly
docker run -p 8080:8080 tw-txparser
```

### Native Go
```bash
./txparser
```

The application will:
1. Connect to the Ethereum RPC endpoint
2. Start backward scanning (if enabled)
3. Begin forward polling for new blocks
4. Start the HTTP server on port 8080

## 📡 API Endpoints

### Subscribe to Address
**POST** `/subscribe`

Subscribe to track transactions for a specific address.

**Request Body:**
```json
{
  "address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6"
}
```

**Response:**
```json
{
  "subscribed": true
}
```

### Get Current Block
**GET** `/current`

Returns the latest processed block number.

**Response:**
```json
{
  "block": 18500000
}
```

### Get Transactions
**GET** `/transactions?address=0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6`

Retrieve all transactions associated with an address.

**Response:**
```json
[
  {
    "hash": "0x1234567890abcdef...",
    "from": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
    "to": "0x8ba1f109551bD432803012645Hac136c",
    "value": "1000000000000000000",
    "block": 18500000
  }
]
```

## 🔍 Parser and Poller Deep Dive

### Parser Component

The **Parser** is the core interface that provides:
- **Address Subscription Management**: Register addresses to track
- **Transaction Retrieval**: Query stored transactions by address
- **Block Tracking**: Monitor the latest processed block

```go
type Parser interface {
    GetCurrentBlock() int
    Subscribe(address string) bool
    GetTransactions(address string) []models.Transaction
}
```

### Poller Component

The **Poller** drives the continuous blockchain monitoring through two distinct phases:

#### 1. Backward Scanning (Historical Data)

When the parser starts, it performs a bounded backward scan to process historical blocks:

```go
// Configuration determines scan behavior
if p.backwardScanEnabled {
    stopAt := latestBlock - p.backwardScanDepth
    if stopAt < 1 {
        stopAt = 1
    }
    go p.scanBackward(ctx, latestBlock-1, stopAt)
}
```

**Key Features:**
- **Bounded Scanning**: Limited by `BACKWARD_SCAN_DEPTH` to prevent excessive processing
- **Concurrent Execution**: Runs in a separate goroutine to avoid blocking
- **Context-Aware**: Respects cancellation signals for graceful shutdown
- **Progress Logging**: Reports progress every 1000 blocks

**Process Flow:**
1. Calculate stop point: `current_block - depth`
2. Iterate from current block down to stop point
3. Process each block and extract transactions
4. Store transactions for both sender and receiver addresses

#### 2. Forward Polling (Real-time Monitoring)

After backward scanning, the parser enters continuous forward polling:

```go
func (p *parserImpl) scanForward(ctx context.Context, ticker *time.Ticker) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            p.checkForNewBlocks()
        }
    }
}
```

**Key Features:**
- **Interval-Based**: Polls every 5 seconds (configurable)
- **Incremental Processing**: Only processes new blocks since last check
- **Error Handling**: Continues operation even if individual block fetches fail
- **State Management**: Maintains current block pointer for efficient processing

**Process Flow:**
1. Query latest block number via `eth_blockNumber`
2. Compare with last processed block
3. Process all new blocks sequentially
4. Update current block pointer

### Transaction Processing

For each block, the parser:

1. **Fetches Block Data**: Uses `eth_getBlockByNumber` with full transaction details
2. **Extracts Transactions**: Iterates through all transactions in the block
3. **Normalizes Data**: Converts hex values to decimal strings
4. **Dual Indexing**: Stores each transaction for both sender and receiver addresses

```go
for _, tx := range block.Transactions {
    // Store for sender
    p.store.AddTransaction(tx.From, models.Transaction{
        Hash:  tx.Hash,
        From:  tx.From,
        To:    tx.To,
        Value: hexToBigIntString(tx.Value),
        Block: number,
    })
    // Store for receiver
    p.store.AddTransaction(tx.To, models.Transaction{
        Hash:  tx.Hash,
        From:  tx.From,
        To:    tx.To,
        Value: hexToBigIntString(tx.Value),
        Block: number,
    })
}
```

### Data Model

Transactions are stored with the following structure:

```go
type Transaction struct {
    Hash  string `json:"hash"`  // Transaction hash
    From  string `json:"from"`  // Sender address
    To    string `json:"to"`    // Receiver address
    Value string `json:"value"` // Amount in wei (decimal string)
    Block int    `json:"block"` // Block number
}
```

## 🧪 Testing

### Native Go Testing
Run the complete test suite:

```bash
go test ./... -v
```

### Docker Testing
Test the Docker image:

```bash
# Build and test the image
docker build -t tw-txparser .

# Run tests in container
docker run --rm tw-txparser go test ./... -v

# Test the running application
docker run -d --name txparser-test -p 8080:8080 tw-txparser
sleep 10
curl http://localhost:8080/current
docker stop txparser-test
```

The test suite includes:
- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end workflow testing
- **Mock Implementations**: Isolated testing of RPC and storage components

## 🔧 Development

### Project Structure

```
tw-txparser/
├── cmd/txparser/          # Main application entry point
├── internal/
│   ├── server/            # HTTP server implementation
│   └── storage/           # In-memory storage implementation
├── pkg/
│   ├── models/            # Domain models
│   ├── parser/            # Parser and poller logic
│   └── rpc/               # Ethereum RPC client
├── Dockerfile             # Multi-stage Docker build
├── docker-compose.yml     # Docker Compose configuration
├── .dockerignore          # Docker ignore file
└── README.md
```

### Key Design Decisions

1. **Interface-Based Architecture**: Clean separation between components
2. **In-Memory Storage**: Fast access for read-heavy workloads
3. **Bounded Backward Scanning**: Prevents excessive historical processing
4. **Dual Transaction Indexing**: Enables efficient address-based queries
5. **Graceful Shutdown**: Proper cleanup on termination signals

## 🚨 Error Handling

The system handles various error scenarios:

- **RPC Failures**: Continues operation, logs errors
- **Invalid Block Data**: Skips problematic blocks
- **Network Issues**: Retries on next polling cycle
- **Storage Errors**: Logs but doesn't crash the application

## 📊 Performance Considerations

- **Memory Usage**: Grows with number of tracked addresses and transactions
- **RPC Rate Limits**: Respects Ethereum node rate limits
- **Concurrent Processing**: Backward and forward scanning run independently
- **Efficient Indexing**: O(1) lookup for address-based queries

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🐳 Docker Features

The Docker setup includes several production-ready features:

### Multi-stage Build
- **Builder Stage**: Uses Go 1.24 Alpine image for compilation
- **Runtime Stage**: Minimal Alpine image for production
- **Size Optimization**: Final image is ~15MB

### Security Features
- **Non-root User**: Runs as `appuser` (UID 1001)
- **Minimal Dependencies**: Only essential packages included
- **No Shell Access**: Reduces attack surface

### Health Monitoring
- **Built-in Health Check**: Monitors `/current` endpoint
- **Graceful Degradation**: Continues operation during temporary failures
- **Container Orchestration**: Compatible with Docker Swarm and Kubernetes

### Environment Configuration
```bash
# Custom configuration
docker run -p 8080:8080 \
  -e BACKWARD_SCAN_ENABLED=false \
  -e BACKWARD_SCAN_DEPTH=5000 \
  tw-txparser
```

## 🆘 Support

For issues and questions:
1. Check the existing issues
2. Create a new issue with detailed description
3. Include logs and configuration details

---

**Note**: This parser is designed for educational and development purposes. For production use, consider additional features like persistent storage, rate limiting, and monitoring.