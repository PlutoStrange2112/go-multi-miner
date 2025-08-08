package multiminer

import (
	"context"
)

// MinerID is a stable identifier for a miner instance.
type MinerID string

// Model describes a miner model/vendor/firmware tuple.
type Model struct {
	Vendor   string // e.g., Bitmain, Whatsminer, Goldshell, iPollo
	Product  string // e.g., S9, M30S, KD-BOX
	Firmware string // e.g., BMminer 2.0, BraiinsOS, VNISH, LuxOS
}

// Endpoint represents how to reach a device.
type Endpoint struct {
	Address string // host:port or http(s)://ip ... depending on driver
}

// Capability declares supported features of a driver/device.
type Capability struct {
	ReadStats           bool
	ReadSummary         bool
	ListPools           bool
	ManagePools         bool // add/enable/disable/remove/switch
	Restart             bool
	Quit                bool
	Commands            []string // optional list of supported raw commands
	FanControl          bool
	PowerControl        bool
	SupportedPowerModes []PowerModeKind
}

// Stats is a generic device metrics snapshot.
type Stats struct {
	Model      Model
	Hashrate5s float64 // GH/s 5s window if available
	HashrateAv float64 // GH/s average
	TempMax    float64
	UptimeSec  int64
}

// Summary is a high-level miner summary.
type Summary struct {
	Accepted              int64
	Rejected              int64
	DeviceHardwarePercent float64
	GHS5s                 float64
	GHSav                 float64
}

// Pool describes a configured pool on device.
type Pool struct {
	ID       int64
	URL      string
	User     string
	Priority int64
	Active   bool
}

// Power/Fan control
type PowerModeKind string

const (
	PowerLow      PowerModeKind = "low"
	PowerBalanced PowerModeKind = "balanced"
	PowerHigh     PowerModeKind = "high"
	PowerCustom   PowerModeKind = "custom"
)

type PowerMode struct {
	Kind      PowerModeKind
	Watts     int // optional target watts
	VoltageMv int // optional millivolts
	FreqMHz   int // optional MHz per chain
}

type FanModeKind string

const (
	FanAuto   FanModeKind = "auto"
	FanManual FanModeKind = "manual"
)

type FanConfig struct {
	Mode     FanModeKind
	SpeedPct int // valid if Mode=manual, 0..100
}

// Driver provides an abstract interface to different miner firmwares.
type Driver interface {
	// Name is a unique identifier for the driver.
	Name() string

	// Detect checks whether the endpoint likely belongs to this driver.
	Detect(ctx context.Context, ep Endpoint) (bool, error)

	// Capabilities returns supported features for the connected endpoint.
	Capabilities() Capability

	// Open creates a session handle to the miner at endpoint.
	Open(ctx context.Context, ep Endpoint) (Session, error)
}

// Session is a connected device handle.
type Session interface {
	Close() error

	Model(ctx context.Context) (Model, error)
	Stats(ctx context.Context) (Stats, error)
	Summary(ctx context.Context) (Summary, error)
	Pools(ctx context.Context) ([]Pool, error)
	AddPool(ctx context.Context, url, user, pass string) error
	EnablePool(ctx context.Context, poolID int64) error
	DisablePool(ctx context.Context, poolID int64) error
	RemovePool(ctx context.Context, poolID int64) error
	SwitchPool(ctx context.Context, poolID int64) error
	Restart(ctx context.Context) error
	Quit(ctx context.Context) error

	// Exec executes an underlying firmware-specific command when supported.
	// The format of parameter depends on driver. For cgminer, it's the Parameter string.
	Exec(ctx context.Context, command string, parameter string) ([]byte, error)

	// Power & fan control (if supported)
	GetPowerMode(ctx context.Context) (PowerMode, error)
	SetPowerMode(ctx context.Context, mode PowerMode) error
	GetFan(ctx context.Context) (FanConfig, error)
	SetFan(ctx context.Context, fan FanConfig) error
}
