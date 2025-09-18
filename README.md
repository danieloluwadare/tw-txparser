# TRUST-WALLET Transaction Parser

A high-performance Ethereum transaction parser that continuously monitors blockchain blocks, extracts transaction data, and provides a REST API for querying transactions by address. The system features intelligent backward scanning, real-time forward polling, and storage abstraction for data persistence.

> **⚠️ PRODUCTION WARNING**: This implementation uses in-memory storage for demonstration purposes only. **Data is lost on restart and memory usage grows indefinitely**. In production environments, transactions **MUST** be stored in a database for persistence, scalability, and reliability.

## 🏗️ Architecture Overview

The system consists of four main components:

- **Parser/Poller**: Core engine that monitors blockchain and processes transactions
- **RPC Client**: Communicates with Ethereum nodes via JSON-RPC
- **Storage**: Abstracted data store for subscriptions and transactions (in-memory for demo, database for production)
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
         └──────────────►│  Data Storage   │
                        │                 │
                        │  - Subscriptions│
                        │  - Transactions │
                        │  (Memory/DB)    │
                        └─────────────────┘
```

## 🚀 Features

- **Real-time Block Monitoring**: Continuously polls for new blocks
- **Historical Data Scanning**: Configurable backward scanning for missed blocks
- **Address Subscription**: Track specific Ethereum addresses
- **Transaction Indexing**: Stores both incoming and outgoing transactions per address (with database persistence in production)
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
| `ETHEREUM_RPC_URL` | `https://ethereum-rpc.publicnode.com` | Ethereum RPC endpoint URL |
| `BACKWARD_SCAN_ENABLED` | `true` | Enable/disable historical block scanning |
| `BACKWARD_SCAN_DEPTH` | `10000` | Number of blocks to scan backward from current |

### Example Configuration

```bash
export ETHEREUM_RPC_URL="https://mainnet.infura.io/v3/YOUR_PROJECT_ID"
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

## 🧪 API Testing with Postman

### 1. Get Current Block - `GET /current`

**Request:**
- **Method:** `GET`
- **URL:** `http://localhost:8080/current`

**Response:**
- **Status:** `200 OK`
- **Response Time:** `2 ms`
- **Response Size:** `136 B`
- **Body:**
```json
{
  "block": 23338992
}
```

### 2. Subscribe to Address - `POST /subscribe`

**Request:**
- **Method:** `POST`
- **URL:** `http://localhost:8080/subscribe`
- **Content-Type:** `application/json`
- **Body:**
```json
{
  "address": "0xa69babef1ca67a37ffaf7a485dfff3382056e78c"
}
```

**Response:**
- **Status:** `200 OK`
- **Response Time:** `1 ms`
- **Response Size:** `137 B`
- **Body:**
```json
{
  "subscribed": true
}
```


### 3. Get Transactions - `GET /transactions`

**Request:**
- **Method:** `GET`
- **URL:** `http://localhost:8080/transactions?address=0xa69babef1ca67a37ffaf7a485dfff3382056e78c`

**Response:**
- **Status:** `200 OK`
- **Response Time:** `2 ms`
- **Response Size:** `53.92 KB`
- **Body:** Array of transaction objects:
```json
[
  {
    "hash": "0xe00269ed013ecc8d90beb5261b6b37587ae9dcf099c5eb30bb439be310da7d61",
    "from": "0x9aab3f81604c683a1a0d14019fbfe15bef7aa1ee",
    "to": "0xa69babef1ca67a37ffaf7a485dfff3382056e78c",
    "value": "9916434",
    "block": 23338991,
    "inbound": true
  },
  {
    "hash": "0x8022594074e7e76ca5678684180deb63cda8a5cf021f3b3d5b8384a836005a2f",
    "from": "0x76dd32063b2899a59f6e15dbc474a160cc922751",
    "to": "0xa69babef1ca67a37ffaf7a485dfff3382056e78c",
    "value": "13630994",
    "block": 23338991,
    "inbound": true
  }
  // ... more transactions
]
```

### Test Results Summary

✅ **All API endpoints are working correctly:**
- Current block tracking is functional (block 23,338,992)
- Address subscription is successful
- Transaction retrieval returns real Ethereum transaction data
- Response times are excellent (< 3ms for all endpoints)
- JSON responses are properly formatted with all required fields

**Note:** The `inbound` field correctly indicates transaction direction - `true` for incoming transactions to the subscribed address.

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

### Storage Interface

The system uses a **storage abstraction** that makes it easy to switch between in-memory and database storage:

```go
type Storage interface {
    Subscribe(address string) bool
    AddTransaction(addr string, tx models.Transaction)
    GetTransactions(address string) []models.Transaction
    IsSubscribed(addr string) bool
}
```

**Current Implementation**: `MemoryStorage` (in-memory)  
**Production Implementation**: Database storage (PostgreSQL, MySQL, etc.)

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

### 🔄 Retry Logic Recommendations

**Note**: The current implementation does not include retry logic for network failures. For production environments, consider implementing the following retry strategies:

#### 1. Exponential Backoff for RPC Calls
```go
// Example retry implementation
func (c *Client) CallWithRetry(ctx context.Context, method string, params []interface{}, result interface{}) error {
    maxRetries := 3
    baseDelay := 1 * time.Second
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := c.Call(ctx, method, params, result)
        if err == nil {
            return nil
        }
        
        // Don't retry on context cancellation
        if ctx.Err() != nil {
            return err
        }
        
        // Exponential backoff
        delay := time.Duration(attempt+1) * baseDelay
        time.Sleep(delay)
    }
    
    return fmt.Errorf("RPC call failed after %d attempts: %w", maxRetries, err)
}
```

#### 2. Circuit Breaker Pattern
Implement a circuit breaker to prevent cascading failures when the RPC endpoint is down:

```go
type CircuitBreaker struct {
    maxFailures int
    timeout     time.Duration
    // ... implementation details
}
```

#### 3. Health Checks and Fallback
- **Health monitoring**: Regular health checks on RPC endpoints
- **Fallback endpoints**: Multiple RPC providers for redundancy
- **Graceful degradation**: Continue with cached data when RPC is unavailable

#### 4. Production Retry Configuration
```go
type RetryConfig struct {
    MaxRetries    int           `json:"max_retries"`
    BaseDelay     time.Duration `json:"base_delay"`
    MaxDelay      time.Duration `json:"max_delay"`
    Jitter        bool          `json:"jitter"`
    RetryableErrors []string    `json:"retryable_errors"`
}
```

**Benefits of implementing retry logic:**
- **Improved reliability**: Handles temporary network issues
- **Better user experience**: Reduces failed requests
- **Production readiness**: Essential for production deployments
- **Cost efficiency**: Reduces unnecessary error handling overhead

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

## 🔄 Data Storage Strategy & Production Considerations

### Current Implementation: Store All Transactions (In-Memory)

**⚠️ IMPORTANT**: This implementation uses **in-memory storage for demonstration purposes only**. In real-world production environments, transactions should be stored in a database for persistence, scalability, and reliability.

The current implementation uses **Option 1: Store All Transactions** approach:

- **All transactions are stored** regardless of subscription status
- **No data loss** when addresses subscribe later
- **Simple implementation** with predictable behavior
- **Memory intensive** but ensures complete historical data
- **⚠️ Data is lost on restart** - not suitable for production

```go
// Current processBlock implementation
for _, tx := range block.Transactions {
    // Store for sender (outbound from their perspective)
    p.store.AddTransaction(tx.From, models.Transaction{
        Hash:    tx.Hash,
        From:    tx.From,
        To:      tx.To,
        Value:   hexToBigIntString(tx.Value),
        Block:   number,
        Inbound: false, // Outbound transaction
    })
    
    // Store for receiver (inbound from their perspective)
    p.store.AddTransaction(tx.To, models.Transaction{
        Hash:    tx.Hash,
        From:    tx.From,
        To:      tx.To,
        Value:   hexToBigIntString(tx.Value),
        Block:   number,
        Inbound: true, // Inbound transaction
    })
}
```

### Alternative Storage Strategies

#### Option 2: Historical Re-scan on Subscription
**Approach**: Only store transactions for subscribed addresses, but re-scan historical blocks when new addresses subscribe.

**Pros**:
- Memory efficient (only relevant data stored)
- No data loss (historical scanning recovers missed transactions)
- Targeted processing

**Cons**:
- More complex implementation
- Requires tracking processed blocks
- Slower subscription process for deeply historical addresses

#### Option 3: Hybrid Approach
**Approach**: Store recent blocks (e.g., last 1000) for all addresses, re-scan older blocks on subscription.

**Pros**:
- Balanced memory usage
- Fast subscription for recent addresses
- Reasonable historical coverage

**Cons**:
- Arbitrary cache size
- Still requires historical scanning for older data


### 🚨 Production Environment Requirements

**CRITICAL**: This in-memory implementation is **NOT suitable for production**. Real-world deployments **MUST** use a database for data persistence.

#### Essential Production Requirements:

1. **Database Storage** (Required)
2. **Data Persistence** (Required) 
3. **Scalability** (Required)
4. **Monitoring** (Required)
5. **Backup & Recovery** (Required)

#### Recommended Production Enhancements:

#### 1. Database Integration
```go
// Example with PostgreSQL
type DatabaseStorage struct {
    db *sql.DB
}

func (d *DatabaseStorage) AddTransaction(addr string, tx models.Transaction) error {
    query := `INSERT INTO transactions (address, hash, from_addr, to_addr, value, block, inbound) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`
    _, err := d.db.Exec(query, addr, tx.Hash, tx.From, tx.To, tx.Value, tx.Block, tx.Inbound)
    return err
}
```

#### 2. Hybrid Approach with Database
- **Recent data** (last N blocks): Store in memory for fast access
- **Historical data**: Store in database with proper indexing
- **Subscription changes**: Query database for historical transactions

#### 3. Performance Optimizations
- **Database indexing** on address and block number
- **Connection pooling** for database access
- **Caching layer** (Redis) for frequently accessed data
- **Batch processing** for bulk operations

#### 4. Scalability Considerations
- **Horizontal scaling** with load balancers
- **Database sharding** by address ranges
- **Message queues** for async processing
- **Monitoring and alerting** for system health

#### 5. Data Retention Policies
- **Configurable retention** periods
- **Archive old data** to cold storage
- **Compression** for historical data
- **Cleanup jobs** for expired subscriptions

### Memory Usage Considerations

**Current Implementation**:
- Stores all transactions for all addresses
- Memory usage grows linearly with blockchain activity
- Suitable for development and small-scale deployments

**Production Scaling**:
- Monitor memory usage patterns
- Implement data retention policies
- Consider database migration when memory becomes a constraint
- Use profiling tools to identify memory hotspots

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