# Software Architecture Analysis: aws-doctor

This document provides a deep-dive architectural analysis of the `aws-doctor` repository. Following the recent decoupling of the monolithic orchestrator, the coordination layer is robust. However, several critical architectural improvements remain in the data presentation, configuration, and observability layers. 

This analysis outlines specific designs and implementation details for resolving these issues.

---

## 1. Generic `WasteItem` Interface (High Priority)

### The Problem
The current design strictly violates the **Open-Closed Principle (OCP)**. `model.RenderWasteInput` is a "god object" containing 20+ separate slices for different AWS resources (e.g., `[]types.Address`, `[]types.Volume`). 

Currently, adding a single new waste check (e.g., "Unused DynamoDB Tables") requires modifying:
1. `model/output.go` (to add the slice to `RenderWasteInput`).
2. `model/output.go` (to add an `if len(src...) > 0 { dest... = append... }` block to the `Merge` function).
3. `utils/waste_table/waste_table.go` (to add a new case to the massive `switch` statement in `RenderScopeTable`).
4. `utils/json_output/json_output.go` (to handle JSON serialization for the new field).

### The Solution
Convert the rigid `RenderWasteInput` structure into a slice of a generic `WasteItem` interface. This allows analyzers to return standardized waste objects that the renderers can display without needing to know their exact concrete types.

**Proposed Interface Design (`model/waste.go`):**
```go
package model

type WasteItem interface {
    // Category returns the high-level category (e.g., "EC2", "RDS", "VPC")
    Category() string
    
    // Issue returns the specific rule violation (e.g., "Unused Elastic IP", "Idle EC2 Instance")
    Issue() string
    
    // ResourceID returns the identifier of the resource (e.g., "i-1234567890abcdef0")
    ResourceID() string
    
    // Location returns the AWS region or global identifier (e.g., "us-east-1")
    Location() string
    
    // EstimatedSavings returns the monthly cost savings if this issue is resolved
    EstimatedSavings(pricingService PricingService) (float64, error)
    
    // AdditionalDetails returns a map of context-specific metadata for rendering (e.g., {"Size": "100GB"})
    AdditionalDetails() map[string]string
}
```

**Refactored `RenderWasteInput` (`model/output.go`):**
```go
type RenderWasteInput struct {
    AccountID string
    Items     []WasteItem
    Errors    map[string]string
}

// The Merge function becomes trivial:
func (dest *RenderWasteInput) Merge(src RenderWasteInput) {
    dest.Items = append(dest.Items, src.Items...)
    for k, v := range src.Errors {
        dest.Errors[k] = v
    }
}
```

**Implementation Steps:**
1. Define the `WasteItem` interface in the `model` package.
2. Implement the `WasteItem` interface on all 20+ existing waste structs (e.g., `EC2IdleInstanceInfo`, `RDSInstanceWasteInfo`).
3. Refactor `RenderWasteInput` to hold `[]WasteItem`.
4. Refactor `utils/waste_table/waste_table.go` to iterate over `Items`, grouping them by `Category()` and dynamically generating table rows based on `AdditionalDetails()`.
5. Update JSON and CSV serializers to iterate over `Items`.

---

## 2. Renderer Interface Abstraction

### The Problem
The presentation logic is scattered. `service/output` acts as a router, but the actual formatting logic resides in loosely grouped utility packages (`utils/tui`, `utils/csv_output`, `utils/json_output`, `utils/waste_table`). This tight coupling to CLI flags inside the service layer makes it difficult to add new output formats (e.g., XML, HTML, Markdown).

### The Solution
Introduce a polymorphic `Renderer` interface that encapsulates all output formats. The `cmd/` package factory will instantiate the correct renderer based on the `--output` flag.

**Proposed Interface Design (`service/output/types.go`):**
```go
type Renderer interface {
    // RenderWaste takes the aggregated waste input and renders it to the user.
    RenderWaste(input model.RenderWasteInput, pricing pricing.Service) error
    
    // RenderCost takes the cost input and renders the comparison.
    RenderCost(input model.RenderCostComparisonInput) error
    
    // RenderTrend renders the 6-month historical trend.
    RenderTrend(input model.TrendInput) error
}
```

**Implementation Steps:**
1. Create `TableRenderer`, `JSONRenderer`, `CSVRenderer`, and `TUIRenderer` structs in `service/output/` that implement the `Renderer` interface.
2. Move the logic from `utils/` directly into these renderer implementations (or have the renderers call the utilities).
3. Inject the `Renderer` interface into the Orchestrator services instead of using the raw `OutputService`.

---

## 3. Structured Configuration Management

### The Problem
The `model.Flags` struct is passed globally across the orchestrators and analyzers. As we add more parameters (e.g., `EC2IdleDays`, `LambdaLookbackDays`), this struct grows linearly. Furthermore, `service_waste.go` currently hardcodes the default thresholds in constants and forcefully overwrites the `Flags` fields. This limits users from overriding these values via CLI flags or config files.

### The Solution
Implement a structured, domain-grouped configuration model, preferably using `spf13/viper` to automatically bind environment variables, config files (`.aws-doctor.yaml`), and CLI flags.

**Proposed Configuration Model:**
```go
type Config struct {
    Output OutputConfig
    EC2    EC2Config
    RDS    RDSConfig
    VPC    VPCConfig
}

type OutputConfig struct {
    Format string // "table", "json", "csv"
    Report bool
}

type EC2Config struct {
    IdleDays               int
    IdleCPUPercent         float64
    IdleNetworkBytesPerDay int
    StoppedDays            int
    // ...
}
```

**Implementation Steps:**
1. Define the hierarchical `Config` struct.
2. Initialize `viper` in `cmd/root.go` and bind it to the Cobra flags.
3. Replace references to `model.Flags` across the codebase with the specific domain configuration (e.g., pass `config.EC2` to `NewEC2Service`).

---

## 4. Structured Logging & Observability

### The Problem
`aws-doctor` currently lacks a leveled logging system. Errors and warnings (like pricing API failures) are written directly to standard streams via `fmt.Fprintf(os.Stderr, ...)`. If users experience a bug or AWS SDK connectivity issues, there is no way for them to enable "debug" mode to view traces or underlying API errors.

### The Solution
Migrate to `log/slog` (the standard Go structured logger).

**Proposed Design:**
1. Add a `--debug` (or `-v`) flag to `cmd/root.go`.
2. Configure `slog` at startup. If `--debug` is true, set the log level to `slog.LevelDebug`.
3. Replace all instances of `fmt.Println`, `fmt.Fprintf(os.Stderr)`, and `log.Print` with `slog.Info()`, `slog.Warn()`, or `slog.Error()`.

**Example Usage:**
```go
if err := s.cfg.PricingService.LoadRegionRates(pricingCtx); err != nil {
    slog.Warn("pricing API partial failure, falling back to defaults", "error", err)
}

// Deep inside analyzers:
slog.Debug("fetching ec2 instances", "region", s.region, "max_results", 100)
```

**Implementation Steps:**
1. Set up the default `slog` handler in `cmd/root.go`.
2. Do a codebase sweep (`grep -r "fmt.Fprintf" .`) to replace hardcoded stderr writes.
3. Add debug trace logs to critical AWS SDK calls to improve the user experience during troubleshooting.
