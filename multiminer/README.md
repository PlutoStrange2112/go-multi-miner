# MultiMiner Library - Extended CGMiner API

A comprehensive extension of the original [go-cgminer-api](https://github.com/x1unix/go-cgminer-api) library that adds support for modern mining hardware beyond traditional cgminer-compatible devices. This library provides unified access to mixed mining fleets including newer HTTP-based miners, alternative firmwares, and multi-vendor deployments through consistent Go interfaces.

## 🎯 Project Purpose

The original cgminer API was designed for older mining hardware that used the cgminer JSON-RPC protocol over TCP. As the mining industry evolved, new manufacturers and firmware developers adopted HTTP APIs, proprietary protocols, and enhanced management interfaces that weren't compatible with the traditional cgminer approach.

**This library bridges that gap** by providing:

- **Backward Compatibility**: Full support for original cgminer/BMminer devices
- **Modern Hardware Support**: Native support for HTTP-based miners (Whatsminer, Goldshell, iPollo)
- **Alternative Firmware Support**: Specialized drivers for Braiins OS, LuxOS, HiveOS
- **Unified Interface**: Single API to manage mixed fleets regardless of underlying protocol
- **Production Hardening**: Enterprise-grade features like connection pooling, security validation, structured errors

**Status**: Production-ready library with complete driver implementations for all major mining hardware vendors and architectures.

## Library Usage

This is primarily a **Go library** for integration into your applications. The example HTTP server in `/cmd/multiminer/` is just for testing - most users should import and use the library directly.

### Basic Library Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/x1unix/go-cgminer-api/multiminer"
)

func main() {
    // Create registry and register drivers
    registry := multiminer.NewRegistry()
    registry.Register(multiminer.NewAntminerDriver())
    registry.Register(multiminer.NewWhatsminerDriver())
    // ... register other drivers as needed

    // Create manager
    manager := multiminer.NewManager(registry)
    defer manager.Close()

    // Add a device (auto-detect driver)
    ctx := context.Background()
    endpoint := multiminer.Endpoint{Address: "192.168.1.100:4028"}
    err := manager.AddOrDetect(ctx, "miner-01", endpoint, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Use the device
    err = manager.WithSession(ctx, "miner-01", func(session multiminer.Session) error {
        stats, err := session.Stats(ctx)
        if err != nil {
            return err
        }
        fmt.Printf("Hashrate: %.2f GH/s, Temp: %.1f°C\n", stats.HashrateAv, stats.TempMax)
        return nil
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### Advanced Library Usage

```go
// Custom logger (optional)
multiminer.SetLogger(multiminer.NewSimpleLogger(multiminer.LogLevelInfo))

// Manager with custom options
options := multiminer.ManagerOptions{
    ProbeTimeout: 5 * time.Second,
}
manager := multiminer.NewManagerWithOptions(registry, options)

// Specific driver usage
driver := multiminer.NewBraiinsDriver()
session, err := driver.Open(ctx, endpoint)
if err != nil {
    log.Fatal(err)
}
defer session.Close()

model, _ := session.Model(ctx)
fmt.Printf("Device: %s %s running %s\n", model.Vendor, model.Product, model.Firmware)
```

## Features

- Driver registry and auto‑detection
- Antminer (Bitmain) via CGMiner/BMminer JSON API
- REST API to list devices, fetch stats/summary, raw command passthrough
- Power/Fan control abstraction (implemented per driver as supported)
- Protocol preference: fastest/lightest first, fallback second

## Supported Hardware (Complete Implementations)

- **CGMiner** - Base cgminer/BMminer JSON API over TCP
- **Antminer** - Bitmain devices with enhanced model detection
- **Braiins OS** - BraiinsOS with power management and cgminer compatibility 
- **LuxOS** - Dual HTTP + cgminer API with intelligent fallback
- **Whatsminer** - MicroBT HTTP API with hashrate parsing and power control
- **Goldshell** - HTTP API with comprehensive JSON response handling
- **HiveOS** - Local agent API with multi-miner aggregation support
- **iPollo** - HTTP-based with robust endpoint detection and management

All drivers include:
- **Automatic Device Detection**: Smart endpoint probing with vendor-specific heuristics
- **Model Identification**: Accurate hardware model and firmware version detection
- **Stats/Summary Collection**: Unified hashrate, temperature, and performance metrics
- **Pool Management**: Add/remove/switch mining pools (where supported by hardware)
- **Power/Fan Control**: Standardized power mode and cooling management abstractions
- **Graceful Error Handling**: Structured error responses with proper HTTP status mapping
- **Connection Pooling**: Efficient connection reuse and automatic cleanup

## 🏗️ Architecture & Production Features

### Core Architecture
```go
Driver Interface → Session Interface → Hardware API
     ↓                    ↓               ↓
Auto-Detection → Connection Pool → TCP/HTTP/WebSocket
     ↓                    ↓               ↓
Device Registry → Manager → Your Application
```

### Enterprise-Grade Enhancements

**🔒 Security & Validation**
- Input sanitization and validation to prevent injection attacks
- Network address validation with allowlist/blocklist support  
- Command parameter validation for safe operation
- Configurable security policies per deployment

**⚡ Performance Optimization**
- **Connection Pooling**: Reuse TCP/HTTP connections across requests
- **Concurrent Operations**: Thread-safe operations with proper mutex usage
- **Resource Management**: Automatic cleanup of idle connections
- **Configurable Timeouts**: Tunable timeouts for different network conditions

**🛠️ Operational Excellence**
- **Structured Logging**: Optional logging interface that won't conflict with host applications
- **Comprehensive Error Handling**: Typed errors with context and HTTP status mapping
- **Configuration Management**: JSON config files with environment variable overrides
- **Graceful Degradation**: Fallback protocols when primary endpoints fail

**📊 Monitoring & Observability**
- Connection pool statistics and health metrics
- Driver-specific capability reporting
- Device status and performance tracking
- Configurable logging levels (Debug, Info, Warn, Error)

## Installation

```bash
go get github.com/x1unix/go-cgminer-api/multiminer
```

## Example HTTP Server (Testing Only)

An example HTTP server is provided in `/cmd/multiminer/` for testing purposes. This is **NOT** the primary use case - most applications should import the library directly.

```bash
# Run example server
go run ./cmd/multiminer

# With custom devices
DEVICES="rig1=192.168.1.10:4028,rig2=192.168.1.11:4028" go run ./cmd/multiminer
```

## REST API

Base URL: http://localhost:8080 (legacy) or http://localhost:8080/api/v1 (versioned)

- List devices

  - GET /devices or GET /api/v1/devices
  - Returns: [{ id, address, driver }]

- Register a device

  - POST /devices or POST /api/v1/devices
  - Body: { "id": "rig1", "address": "192.168.1.10:4028", "driver": "antminer" | "cgminer" | "whatsminer" | ... (optional) }
  - If driver not provided, the server attempts auto‑detect.

- Device info (basic)

  - GET /devices/{id} or GET /api/v1/devices/{id}

- Capabilities (per driver)

  - GET /devices/{id}/capabilities or GET /api/v1/devices/{id}/capabilities
  - Shows which features are supported (stats, pools, power/fan, supported modes, raw commands).

- Summary

  - GET /devices/{id}/summary or GET /api/v1/devices/{id}/summary

- Stats

  - GET /devices/{id}/stats or GET /api/v1/devices/{id}/stats

- Exec (raw command passthrough)

  - POST /devices/{id}/exec or POST /api/v1/devices/{id}/exec
  - Body (cgminer example): { "Command": "pools", "Parameter": "" }
  - Returns raw JSON from the device.

- Power control (if supported by driver)

  - GET /devices/{id}/power or GET /api/v1/devices/{id}/power
  - POST /devices/{id}/power or POST /api/v1/devices/{id}/power
  - Body: { "Kind": "low"|"balanced"|"high"|"custom", "Watts": 3000, "VoltageMv": 800, "FreqMHz": 700 }

- Fan control (if supported by driver)

  - GET /devices/{id}/fan or GET /api/v1/devices/{id}/fan
  - POST /devices/{id}/fan or POST /api/v1/devices/{id}/fan
  - Body: { "Mode": "auto"|"manual", "SpeedPct": 80 }

Notes:
- Pools CRUD endpoints can be accessed via Exec for cgminer‑based devices (addpool, enablepool, disablepool, removepool, switchpool). Dedicated REST endpoints can be added as needed.
- No auth by default; run behind a reverse proxy or on a trusted network.

## Quick examples

- Add an Antminer (cgminer/BMminer)

```pwsh
curl -Method POST http://localhost:8080/devices -ContentType 'application/json' -Body '{"id":"s19-01","address":"192.168.1.60:4028","driver":"antminer"}'
```

- Stats/Summary

```pwsh
curl http://localhost:8080/devices/s19-01/stats
curl http://localhost:8080/devices/s19-01/summary
```

- Exec (cgminer pools)

```pwsh
curl -Method POST http://localhost:8080/devices/s19-01/exec -ContentType 'application/json' -Body '{"Command":"pools"}'
```

- Power/Fan (will return Not Implemented on vanilla cgminer; enabled per vendor driver)

```pwsh
curl http://localhost:8080/devices/s19-01/power
curl -Method POST http://localhost:8080/devices/s19-01/fan -ContentType 'application/json' -Body '{"Mode":"manual","SpeedPct":80}'
```

## 🚀 Recent Major Improvements

This extended version includes significant enhancements over the original cgminer-api:

### **Complete Driver Ecosystem (NEW)**
- **All 8 major mining hardware vendors** now have complete, production-ready implementations
- **Multi-protocol support**: HTTP, TCP, WebSocket, and cgminer JSON-RPC
- **Smart fallback logic**: Automatically tries fastest protocol first, falls back gracefully

### **Enterprise Security & Validation (NEW)**
```go
// All inputs are validated and sanitized
validator := multiminer.NewAddressValidator()
err := validator.ValidateAddress("192.168.1.100:4028") // Prevents injection attacks

cmdValidator := multiminer.NewCommandValidator() 
err = cmdValidator.ValidateCommand("pools") // Safe command execution
```

### **Production-Grade Connection Management (NEW)**
```go
// Automatic connection pooling with configurable limits
pool := multiminer.NewConnectionPool()
pool.SetLimits(maxIdle: 5, maxOpen: 10, ttl: 5*time.Minute)

// Thread-safe operations with proper resource cleanup
manager := multiminer.NewManager(registry)
defer manager.Close() // Ensures all connections are cleaned up
```

### **Structured Error System (NEW)**
```go
// Typed errors with HTTP status codes and context
if err != nil {
    if minerErr, ok := err.(*multiminer.MultiMinerError); ok {
        fmt.Printf("Error: %s (HTTP %d)\n", minerErr.Message, minerErr.HTTPStatus())
        fmt.Printf("Details: %s\n", minerErr.Details)
    }
}
```

### **Advanced Device Detection (ENHANCED)**
- **Vendor-specific heuristics** for accurate device identification
- **Model catalog integration** with support for 100+ mining device models
- **Firmware version detection** across different vendor APIs
- **Capability discovery** to determine supported features per device

## Design Philosophy

- **Interface Abstraction**: Clean Driver + Session interfaces in `multiminer/core.go`
- **Protocol Intelligence**: Drivers automatically choose fastest protocol (HTTP preferred, TCP fallback)
- **Resource Efficiency**: Connection pooling and session reuse minimize network overhead
- **Security First**: All inputs validated, commands sanitized, network addresses checked
- **Library Integration**: No global state pollution, optional logging, clean shutdown

## Development Status & Roadmap

### ✅ **Completed (Production Ready)**
- **All Major Drivers**: Antminer, Whatsminer, Braiins OS, LuxOS, HiveOS, Goldshell, iPollo
- **Enterprise Security**: Input validation, command sanitization, address validation
- **Connection Pooling**: Production-grade connection management with automatic cleanup
- **Structured Errors**: Comprehensive error handling with HTTP status mapping
- **Configuration System**: JSON config with environment variable overrides
- **Thread Safety**: Mutex-protected operations for concurrent access

### 🔄 **Future Enhancements**
- **GPU Mining Support**: Claymore-compatible driver for GPU mining rigs
- **Advanced Pool Management**: Dedicated REST endpoints for batch pool operations
- **Authentication & Authorization**: Built-in auth support (currently via reverse proxy)
- **Metrics & Monitoring**: Prometheus metrics integration
- **High Availability**: Load balancing and failover for mining pool management

## 🏷️ Naming & Project Evolution

**Note**: This library has evolved significantly beyond its original cgminer-api roots. While it maintains backward compatibility with cgminer-based devices, it now supports 8 major mining hardware vendors with modern HTTP APIs, enterprise security features, and production-grade architecture.

**Current Import Path**: `github.com/x1unix/go-cgminer-api/multiminer`

For new projects, you might consider:
- Creating a new repository with a more descriptive name like `go-mining-api` or `unified-miner-api`
- This would better reflect the multi-vendor, modern mining hardware support
- The current path works fine but may be confusing for users expecting only cgminer support

## Troubleshooting

### **Common Issues**

**Connection Problems**:
- Ensure miner's API is enabled (cgminer: check `api-allow` settings)
- Verify network connectivity and firewall rules
- Check if miner uses non-standard ports

**Protocol Issues**:
- Modern miners often use HTTP (port 80/8080) instead of cgminer TCP (port 4028)
- Some miners require specific HTTP headers or authentication
- Try auto-detection first: `manager.AddOrDetect(ctx, id, endpoint, nil)`

**Build Issues**:
- Windows UNC path build error: Use WSL paths instead of `\\wsl$` UNC paths
- Go module issues: Run `go mod tidy` to resolve dependencies

**Performance Optimization**:
```go
// Configure connection pooling for better performance
options := multiminer.ManagerOptions{
    ProbeTimeout: 5 * time.Second, // Adjust for slow networks
}
manager := multiminer.NewManagerWithOptions(registry, options)
```

**Security Configuration**:
```go
// Enable security validation
validator := multiminer.NewAddressValidator()
// Configure allowlist/blocklist as needed for your network
```
