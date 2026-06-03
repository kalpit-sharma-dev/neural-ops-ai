# Architecture (summary)

NeuralOps is an AI-powered observability platform. This file is a **short index**; the full architecture reference for the architecture team lives in:

**[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)**

That document includes:

- C4 system context and container diagrams  
- Deployment (Docker Compose & Kubernetes)  
- Backend microservices and gateway component model  
- Observability UI API layer (route domains)  
- Frontend architecture and auth flow  
- PostgreSQL ER diagrams (core, FinOps, UI entities)  
- Section-wise sequence diagrams (ingest, search, incidents, alerts, FinOps, AI, collectors, streaming)  
- Security and platform observability  

## Quick reference

```
Clients → Frontend (:3000) → API Gateway (:8080)
                ├── Observability UI APIs (/api/v1/* on gateway)
                └── Proxies → Ingestion, Search, Incident, Analysis, Alerting
Data: PostgreSQL, ClickHouse, Elasticsearch, Kafka, Redis, Qdrant
```

| Service | Port |
|---------|------|
| Frontend | 3000 |
| Gateway | 8080 |
| Ingestion | 8081 |
| Analysis | 8082 |
| Correlation | 8083 |
| Incident | 8084 |
| Search | 8085 |
| Alerting | 8086 |

See [README.md](README.md) for quick start commands.
