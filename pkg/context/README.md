# Context Package

The `context` package provides core functionality for generating standardized resource names, tags, and metadata for cloud infrastructure. This package can be used standalone in Go applications such as serverless functions, CLI tools, or custom automation.

## Installation

```bash
go get github.com/kbrockhoff/terraform-provider-context/pkg/context
```

## Features

- **Name Generation**: Create standardized, cloud-compliant resource name prefixes
- **Tag Generation**: Generate comprehensive resource tags with cloud provider-specific formatting
- **Git Integration**: Automatically include repository and commit information
- **Cloud Provider Support**: AWS, Azure, GCP, and other cloud providers
- **Multiple Output Formats**: Convert tags to various formats (maps, lists, CSV, key=value pairs)

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/kbrockhoff/terraform-provider-context/pkg/context"
)

func main() {
    // Generate a name prefix
    nameGen := &context.NameGenerator{
        Namespace:   "myorg",
        Name:        "api",
        Environment: "prod",
    }
    namePrefix, err := nameGen.Generate()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Name:", namePrefix) // Output: myorg-api-prod

    // Generate tags
    config := &context.DataSourceConfig{
        Namespace:             "myorg",
        Environment:           "prod",
        EnvironmentName:       "Production",
        Availability:          "standard",
        ManagedBy:             "terraform",
        SourceRepoTagsEnabled: true,
        NotApplicableEnabled:  true,
        AdditionalTags:        make(map[string]string),
        AdditionalDataTags:    make(map[string]string),
    }

    cloudProvider := context.GetCloudProvider("aws")
    tagProcessor := &context.TagProcessor{
        CloudProvider: cloudProvider,
        Config:        config,
        TagPrefix:     "app-",
    }

    tags, err := tagProcessor.Process()
    if err != nil {
        log.Fatal(err)
    }

    // Print tags
    for k, v := range tags {
        fmt.Printf("%s = %s\n", k, v)
    }
}
```

## API Reference

### Name Generation

#### NameGenerator

```go
type NameGenerator struct {
    Namespace   string // Organization/team namespace (1-8 chars)
    Name        string // Resource name
    Environment string // Environment identifier (1-8 chars)
}
```

**Methods:**
- `Generate() (string, error)`: Generates a standardized name prefix

**Example:**
```go
gen := &context.NameGenerator{
    Namespace:   "platform",
    Name:        "payment-service",
    Environment: "prod",
}
name, err := gen.Generate()
// Result: "platform-payment-service-prod"
```

### Tag Generation

#### TagProcessor

```go
type TagProcessor struct {
    CloudProvider CloudProvider
    Config        *DataSourceConfig
    TagPrefix     string
}
```

**Methods:**
- `Process() (map[string]string, error)`: Generates main resource tags
- `ProcessDataTags() (map[string]string, error)`: Generates data classification tags

#### DataSourceConfig

Configuration for tag generation:

```go
type DataSourceConfig struct {
    // Naming
    Namespace       string
    Name            string
    Environment     string
    EnvironmentName string
    EnvironmentType string // None, Ephemeral, Development, Testing, UAT, Production, MissionCritical

    // Resource Management
    Enabled      bool
    Availability string   // preemptable, spot, standard, dedicated, isolated
    ManagedBy    string
    DeletionDate string

    // Integration
    PMPlatform      string
    PMProjectCode   string
    ITSMPlatform    string
    ITSMSystemID    string
    ITSMComponentID string
    ITSMInstanceID  string

    // Ownership
    CostCenter    string
    ProductOwners []string
    CodeOwners    []string
    DataOwners    []string

    // Data Classification
    Sensitivity    string   // public, internal, confidential, restricted, critical
    DataRegs       []string // GDPR, CCPA, etc.
    SecurityReview string
    PrivacyReview  string

    // Feature Toggles
    SourceRepoTagsEnabled bool // Include git repository tags
    SystemPrefixesEnabled bool // Add platform prefixes to system IDs
    NotApplicableEnabled  bool // Include N/A for empty values
    OwnerTagsEnabled      bool // Include owner tags

    // Additional Tags
    AdditionalTags     map[string]string
    AdditionalDataTags map[string]string
}
```

### Cloud Provider Support

#### CloudProvider Interface

```go
type CloudProvider interface {
    GetMaxTagLength() int
    GetDelimiter() string
    GetNAValue() string
    SanitizeTagValue(value string) string
    ValidateTagKey(key string) bool
}
```

**Supported Providers:**
- `aws`: Amazon Web Services
- `az`: Microsoft Azure
- `gcp`: Google Cloud Platform
- `dc`: Default/Data Center (generic)
- `oci`, `ibm`, `do`, `vul`, `ali`, `cv`: Other cloud providers

**Example:**
```go
awsProvider := context.GetCloudProvider("aws")
// AWS: 256 char limit, space delimiter, "N/A" value

gcpProvider := context.GetCloudProvider("gcp")
// GCP: 63 char limit, underscore delimiter, "not_applicable" value
```

### Utility Functions

#### Tag Conversion

Convert tags to different formats:

```go
// Convert to AWS-style list of maps
func ConvertTagsToListOfMaps(tags map[string]string) []map[string]string

// Convert to key=value pairs
func ConvertTagsToKVPList(tags map[string]string) []string

// Convert to comma-separated string
func ConvertTagsToCommaSeparated(tags map[string]string) string
```

**Example:**
```go
tags := map[string]string{"env": "prod", "team": "platform"}

// For AWS resources
awsTags := context.ConvertTagsToListOfMaps(tags)
// Result: [{"key": "env", "value": "prod"}, {"key": "team", "value": "platform"}]

// For Docker/container labels
kvpTags := context.ConvertTagsToKVPList(tags)
// Result: ["env=prod", "team=platform"]

// For logging/display
csvTags := context.ConvertTagsToCommaSeparated(tags)
// Result: "env=prod,team=platform"
```

### Git Integration

#### GitInfo

```go
type GitInfo struct {
    RepoURL    string // Repository URL (converted to HTTPS)
    CommitHash string // Full commit hash
}
```

**Functions:**
```go
// Get repository information (cached for 5 minutes)
func GetGitInfo() (*GitInfo, error)

// Clear the cache
func ClearGitCache()
```

**Example:**
```go
gitInfo, err := context.GetGitInfo()
if err == nil {
    fmt.Println("Repo:", gitInfo.RepoURL)
    fmt.Println("Commit:", gitInfo.CommitHash)
}
```

### Validation

Validation functions for input values:

```go
func ValidateNamespace(namespace string) error
func ValidateEnvironment(environment string) error
func ValidateCloudProvider(provider string) error
func ValidateEnvironmentType(envType string) error
func ValidateAvailability(availability string) error
func ValidateSensitivity(sensitivity string) error
func ValidateDeletionDate(date string) error
func ValidateEmail(email string) error
func ValidateEmails(emails []string) error
```

## Use Cases

### Serverless Functions

Use this package in AWS Lambda, Azure Functions, or Google Cloud Functions to generate consistent resource metadata:

```go
import "github.com/kbrockhoff/terraform-provider-context/pkg/context"

func handler() {
    tagProcessor := &context.TagProcessor{
        CloudProvider: context.GetCloudProvider("aws"),
        Config: &context.DataSourceConfig{
            Namespace:   "data",
            Environment: "prod",
            ManagedBy:   "lambda",
        },
        TagPrefix: "app-",
    }

    tags, _ := tagProcessor.Process()
    // Use tags when creating resources
}
```

### CLI Tools

Build command-line tools that generate infrastructure metadata:

```go
// cli/main.go
import "github.com/kbrockhoff/terraform-provider-context/pkg/context"

func main() {
    // Parse flags, generate names and tags, output results
}
```

### Custom Automation

Integrate into custom automation workflows:

```go
import "github.com/kbrockhoff/terraform-provider-context/pkg/context"

func createResource(name string) error {
    // Generate standardized name
    nameGen := &context.NameGenerator{
        Namespace: "auto",
        Name:      name,
        Environment: "prod",
    }
    resourceName, _ := nameGen.Generate()

    // Generate tags
    // ...

    // Create resource with name and tags
    return nil
}
```

## Examples

See the [examples/client-app](../../examples/client-app) directory for a complete working example.

## License

See the repository root for license information.
