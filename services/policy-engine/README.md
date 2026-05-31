## Policy Engine

The Policy Engine is the core compliance and risk evaluation service that classifies transactions and determines their routing path.

### Features

- **Risk Classification**: L1-L4 risk levels based on sender/receiver scores and transaction data
- **Routing Decisions**: Determines routing path (PATH_A/B/C or BLOCKED)
- **OPA Integration**: Uses Open Policy Agent for flexible, auditable policy evaluation
- **High Performance**: 10,000+ TPS capacity
- **Prometheus Metrics**: Built-in observability

### Architecture

```
Transaction Input
    ↓
OPA Policy Engine
    ├─ Risk Classification (L1-L4)
    ├─ Routing Path Decision (PATH_A/B/C/BLOCKED)
    └─ Approval Logic
    ↓
PolicyEvaluation Output
```

### Risk Levels

- **L1 (Green)**: Safe transactions, sender_score < 40, receiver_score < 40
  - Routing: PATH_A (Direct settlement)
  - Action: Automatic approval

- **L2 (Orange)**: Enhanced verification needed, 40 ≤ scores < 60
  - Routing: PATH_B (Enhanced checks)
  - Action: Bank review, insurance coverage

- **L3 (Red)**: Manual review required, 60 ≤ scores < 80
  - Routing: PATH_C (Manual approval)
  - Action: Legal team review, multi-signature

- **L4 (Dark Red)**: Blocked, scores ≥ 80 or sanctions match
  - Routing: BLOCKED
  - Action: Automatic rejection

### API Endpoints

#### Evaluate Transaction
```
POST /evaluate
Content-Type: application/json

{
  "id": "txn-123",
  "amount": 50000,
  "currency": "USD",
  "sender_country": "US",
  "receiver_country": "GB",
  "transaction_type": "wire_transfer",
  "sender_risk_score": 25,
  "receiver_risk_score": 15,
  "metadata": {}
}

Response:
{
  "transaction_id": "txn-123",
  "risk_level": "L1",
  "routing_path": "PATH_A",
  "approved": true,
  "reason": "Low risk transaction",
  "timestamp": 1234567890
}
```

#### List Policies
```
GET /policies

Response:
{
  "policies": [
    {
      "name": "risk_classification",
      "description": "Classify transaction risk level (L1-L4)",
      "version": "1.0.0"
    },
    ...
  ]
}
```

#### Health Check
```
GET /health

Response: {"status":"ok","service":"policy-engine"}
```

### Usage

#### Docker Compose
```bash
docker-compose up policy-engine
```

#### Local Development
```bash
cd services/policy-engine
go run main.go
```

#### Kubernetes (Helm)
```bash
helm install sovereign-trade ./helm/sovereign-trade
```

### Configuration

Environment variables:
- `PORT`: HTTP server port (default: 8080)
- `LOG_LEVEL`: Zap log level (default: info)
- `OPA_POLICY_PATH`: Path to OPA policy files (default: ./opa)

### Performance

- **Throughput**: 10,000+ TPS
- **Latency**: < 10ms (p99)
- **Memory**: ~100MB per instance
- **CPU**: 100-500m per instance

### Monitoring

Prometheus metrics available at `/metrics`:
- `policy_engine_evaluations_total`: Total evaluations
- `policy_engine_evaluation_duration_seconds`: Evaluation latency
- `policy_engine_risk_levels`: Distribution by risk level
- `policy_engine_routing_paths`: Distribution by routing path

### Testing

```bash
# Unit tests
go test ./...

# Integration tests
docker-compose -f docker-compose.yml up -d
go test -tags=integration ./...
```

### Dependencies

- Open Policy Agent (OPA): Policy evaluation
- Chi: HTTP router
- Prometheus: Metrics
- Zap: Structured logging
