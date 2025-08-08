# Go Multi-Miner API

[![MIT Licence](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/mit-license)

A comprehensive, production-ready Go library for managing mixed cryptocurrency mining fleets. This enhanced fork extends the original [go-cgminer-api](https://github.com/x1unix/go-cgminer-api) by [x1unix](https://github.com/x1unix) with critical bug fixes, modern mining hardware support, and enterprise-grade features.

## 🙏 Credits

This project builds upon the excellent work of:
- **[x1unix/go-cgminer-api](https://github.com/x1unix/go-cgminer-api)** - The primary foundation providing cgminer API support and multi-miner architecture
- **[crypt0train/go-cgminer-api](https://github.com/crypt0train/go-cgminer-api)** - The original cgminer API implementation

## 🚀 Key Improvements in This Fork

### **Critical Bug Fixes**
- ✅ **Fixed JSON unmarshaling failures** in TestStatsS9, TestStatsL3plus, TestStatsD3, TestStatsT9
- ✅ **Resolved `,string` struct tag issues** that caused parsing errors with Go 1.14+
- ✅ **Enhanced Number type** to handle both quoted strings and numeric JSON values
- ✅ **Verified ConnectionPool** implementation works correctly

### **Production Enhancements**  
- 🔧 **Modernized build system** from deprecated `dep` to Go modules
- 🏗️ **Improved Makefile** with comprehensive build, test, and install targets
- 🧪 **All core tests now pass** - reliable JSON parsing across all miner types
- 📝 **Updated documentation** with clear usage examples

## 🎯 What This Library Does

The original cgminer API was designed for older mining hardware using JSON-RPC over TCP. As the mining industry evolved, manufacturers adopted HTTP APIs and proprietary protocols. **This library bridges that gap** by providing:

- **🔄 Backward Compatibility**: Full support for original cgminer/BMminer devices  
- **🚀 Modern Hardware Support**: Native support for HTTP-based miners (Whatsminer, Goldshell, iPollo)
- **⚡ Alternative Firmware Support**: Specialized drivers for Braiins OS, LuxOS, HiveOS
- **🎯 Unified Interface**: Single API to manage mixed fleets regardless of protocol
- **🏢 Enterprise Features**: Connection pooling, security validation, structured errors

## 📦 Installation

```bash
go get github.com/PlutoStrange2112/go-multi-miner
```

## 🔧 Supported Hardware (Complete Implementations)

| Hardware | Protocol | Status | Features |
|----------|----------|---------|----------|
| **CGMiner** | TCP JSON-RPC | ✅ Complete | Base cgminer/BMminer compatibility |
| **Antminer (Bitmain)** | TCP JSON-RPC | ✅ Complete | Enhanced model detection, S9/S19/D3/L3+/T9+ |
| **Braiins OS** | TCP/HTTP | ✅ Complete | Power management, cgminer compatibility |
| **LuxOS** | HTTP + TCP | ✅ Complete | Dual protocol with intelligent fallback |
| **Whatsminer (MicroBT)** | HTTP API | ✅ Complete | M30S+, M50S+, hashrate parsing, power control |
| **Goldshell** | HTTP API | ✅ Complete | KD-BOX, HS-BOX series with JSON handling |
| **HiveOS** | HTTP API | ✅ Complete | Local agent API, multi-miner aggregation |
| **iPollo** | HTTP API | ✅ Complete | V1/G1 series with endpoint detection |

## 🚀 Quick Start

### Basic Usage (Single Device)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/PlutoStrange2112/go-multi-miner"
)

func main() {
    // Create registry and register drivers
    registry := multiminer.NewRegistry()
    registry.Register(multiminer.NewAntminerDriver())
    registry.Register(multiminer.NewWhatsminerDriver())
    // Add other drivers as needed...

    // Create manager with connection pooling
    manager := multiminer.NewManager(registry)
    defer manager.Close()

    // Add a device (auto-detect driver)
    ctx := context.Background()
    endpoint := multiminer.Endpoint{Address: "192.168.1.100:4028"}
    err := manager.AddOrDetect(ctx, "miner-01", endpoint, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Get device statistics
    err = manager.WithSession(ctx, "miner-01", func(session multiminer.Session) error {
        stats, err := session.Stats(ctx)
        if err != nil {
            return err
        }
        
        fmt.Printf("Miner: %s %s\n", stats.Model.Vendor, stats.Model.Product)
        fmt.Printf("Hashrate: %.2f GH/s (5s: %.2f GH/s)\n", stats.HashrateAv, stats.Hashrate5s)
        fmt.Printf("Temperature: %.1f°C\n", stats.TempMax)
        fmt.Printf("Uptime: %d seconds\n", stats.UptimeSec)
        return nil
    })
    
    if err != nil {
        log.Fatal(err)
    }
}
```

### Advanced Usage (Fleet Management)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "github.com/PlutoStrange2112/go-multi-miner"
)

func main() {
    // Configure logging (optional)
    multiminer.SetLogger(multiminer.NewSimpleLogger(multiminer.LogLevelInfo))

    // Create registry with all drivers
    registry := multiminer.NewRegistry()
    drivers := []multiminer.Driver{
        multiminer.NewAntminerDriver(),
        multiminer.NewWhatsminerDriver(), 
        multiminer.NewBraiinsDriver(),
        multiminer.NewLuxOSDriver(),
        multiminer.NewGoldshellDriver(),
        multiminer.NewHiveOSDriver(),
        multiminer.NewIPolloDriver(),
    }
    
    for _, driver := range drivers {
        registry.Register(driver)
    }

    // Create manager with custom options
    options := multiminer.ManagerOptions{
        ProbeTimeout: 10 * time.Second,
    }
    manager := multiminer.NewManagerWithOptions(registry, options)
    defer manager.Close()

    ctx := context.Background()
    
    // Add multiple devices
    miners := map[string]string{
        "antminer-s19": "192.168.1.100:4028",
        "whatsminer-m30": "192.168.1.101:8080", 
        "goldshell-kd": "192.168.1.102:80",
    }

    for id, address := range miners {
        endpoint := multiminer.Endpoint{Address: address}
        if err := manager.AddOrDetect(ctx, multiminer.MinerID(id), endpoint, nil); err != nil {
            log.Printf("Failed to add %s: %v", id, err)
            continue
        }
        fmt.Printf("✅ Added %s at %s\n", id, address)
    }

    // Get fleet statistics
    devices := manager.List()
    for _, device := range devices {
        err := manager.WithSession(ctx, device.ID, func(session multiminer.Session) error {
            model, _ := session.Model(ctx)
            stats, err := session.Stats(ctx)
            if err != nil {
                return err
            }
            
            fmt.Printf("\n📊 %s (%s %s)\n", device.ID, model.Vendor, model.Product)
            fmt.Printf("   Hashrate: %.2f GH/s\n", stats.HashrateAv)
            fmt.Printf("   Temperature: %.1f°C\n", stats.TempMax)
            
            return nil
        })
        
        if err != nil {
            log.Printf("❌ Failed to get stats for %s: %v", device.ID, err)
        }
    }
}
```

### Legacy CGMiner Usage

For backward compatibility with the original cgminer API:

```go
package main

import (
    cgminer "github.com/PlutoStrange2112/go-multi-miner"
    "time"
    "log"
    "fmt"
)

func main() {
    miner := cgminer.NewCGMiner("192.168.1.100", 4028, 5*time.Second)
    
    // Get generic stats
    stats, err := miner.Stats()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Type: %s | GHS avg: %.2f\n", stats.Type, stats.GhsAverage)
    
    // Get device-specific stats
    if statsS9, err := stats.S9(); err == nil {
        fmt.Printf("S9 Frequency: %.2f MHz\n", statsS9.Frequency.Float64())
    }
    
    if statsL3, err := stats.L3(); err == nil {
        fmt.Printf("L3+ Hashrate: %.2f GH/s\n", statsL3.Ghs5s.Float64())
    }
}
```

## 🏗️ Architecture

```
Application Layer
       ↓
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│    Manager      │ ←→ │   Registry      │ ←→ │ Connection Pool │
└─────────────────┘    └─────────────────┘    └─────────────────┘
       ↓                        ↓                        ↓
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Driver        │ ←→ │    Session      │ ←→ │   Transport     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
       ↓                        ↓                        ↓
┌──────────────────────────────────────────────────────────────────┐
│              Hardware APIs (TCP/HTTP/WebSocket)                   │
└──────────────────────────────────────────────────────────────────┘
```

## 🧪 Building and Testing

```bash
# Install dependencies
make deps

# Run all tests  
make test

# Run multiminer tests specifically
make multiminer-test

# Generate coverage report
make cover

# Build the example server
make build

# Clean build artifacts
make clean
```

## 🔧 HTTP Server Example

An example HTTP server is included for testing:

```bash
# Run with default configuration
go run ./cmd/multiminer

# Run with custom listen address
go run ./cmd/multiminer -listen :9090

# Load devices from environment
DEVICES="rig1=192.168.1.10:4028,rig2=192.168.1.11:8080" go run ./cmd/multiminer
```

### REST API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/devices` | GET | List all devices |
| `/api/v1/devices` | POST | Register new device |
| `/api/v1/devices/{id}` | GET | Get device info |
| `/api/v1/devices/{id}/stats` | GET | Get device statistics |
| `/api/v1/devices/{id}/summary` | GET | Get device summary |
| `/api/v1/devices/{id}/pools` | GET | List mining pools |
| `/api/v1/devices/{id}/exec` | POST | Execute raw command |

## 🔒 Security Features

- **Input Validation**: Sanitized parameters prevent injection attacks
- **Address Validation**: Network address allowlist/blocklist support  
- **Command Validation**: Restricted command execution for safety
- **Connection Limits**: Configurable connection pool limits
- **Timeout Protection**: Configurable timeouts prevent hanging operations

## 📈 Production Features

- **Connection Pooling**: Efficient connection reuse and management
- **Structured Errors**: HTTP-compatible error codes and messages
- **Graceful Shutdown**: Clean resource cleanup on termination
- **Background Cleanup**: Automatic expired connection removal
- **Comprehensive Logging**: Structured logging with configurable levels
- **Thread Safety**: Mutex-protected concurrent operations

## 🔄 Migration from Original

If migrating from `x1unix/go-cgminer-api`:

1. Update import path:
   ```go
   // Old
   import "github.com/x1unix/go-cgminer-api"
   
   // New  
   import "github.com/PlutoStrange2112/go-multi-miner"
   ```

2. The cgminer API remains backward compatible
3. All existing cgminer functionality works as expected
4. New multiminer features are opt-in

## 📋 Breaking Changes from Original

- `cgminer.NewCGMiner()` accepts `time.Duration` instead of `int` seconds
- Some fields changed from `json.Number` to `cgminer.Number` for better JSON handling
- Enhanced error types with structured information

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)  
5. Open a Pull Request

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [x1unix/go-cgminer-api](https://github.com/x1unix/go-cgminer-api) - Original foundation library
- [crypt0train/go-cgminer-api](https://github.com/crypt0train/go-cgminer-api) - Initial cgminer implementation

---

**Made with ❤️ for the cryptocurrency mining community**