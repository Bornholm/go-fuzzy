# Fuzzy Server

A very simple HTTP server exposing fuzzy engines as a JSON API.

## Getting started

### Running the server

```bash
# Start the server and load engines definition from the cmd/fuzzy/examples directory
go run ./cmd/fuzzy -files './cmd/fuzzy/examples/*.fuzzy'
```

## API

### `GET /api/v1/engines`

List loaded engine definitions.

### `GET /api/v1/engines/{name}`

Retrieve the given named engine definition as its JSON representation.

### `POST /api/v1/engines/{name}`

Send values to compute to the named engine.

**cURL Example**

```bash
curl -d '{"resource_availability":50,"response_time_trend":0,"pod_count":8}' 'http://localhost:3003/api/v1/engines/pod-autoscaler'
```
