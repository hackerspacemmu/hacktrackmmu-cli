# HacktrackMMU CLI

cli tool for hacktrack cuz why not

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install) (1.21+ recommended, currently configured for `latest` via `mise`)
- `make` (optional, for running task automation)

### Installation

#### Quick Install (via curl)

You can install `hacktrackmmu-cli` using the following curl command:

```bash
curl -fsSL https://raw.githubusercontent.com/hackerspacemmu/hacktrackmmu-cli/main/install.sh | bash
```

#### From Source

1. Clone or copy this repository to your local machine:
   ```bash
   git clone https://github.com/hackerspacemmu/hacktrackmmu-cli.git
   cd hacktrackmmu-cli
   ```
2. Initialize and tidy dependencies:
   ```bash
   go mod tidy
   ```

---

## Building and Running

You can use the provided `Makefile` to simplify development tasks:

### Tasks

- **Build the binary**:
  ```bash
  make build
  ```
  This places the compiled binary in `bin/hacktrackmmu-cli` with injected build information.

- **Run the CLI directly**:
  ```bash
  make run ARGS="version"
  # Or with flags:
  make run ARGS="version -v"
  ```

- **Run tests**:
  ```bash
  make test
  ```

- **Format code**:
  ```bash
  make fmt
  ```

- **Install to global `GOBIN`**:
  ```bash
  make install
  ```

---

## Configuration

The CLI load order for configurations is:
1. Local flags (e.g. `--verbose`)
2. Environment Variables (prefixed with `HACKTRACK_`)
3. Configuration file (defaults to `config.yaml` in the working directory, or `$HOME/.config/hacktrackmmu-cli/config.yaml`)
4. Struct Defaults

### Configuration File Template

Create a `config.yaml` file in the root of the project to customize settings:

```yaml
# config.yaml
log_level: info
api_url: "api-url"
```

---

## Usage Guide

### Root Command
```bash
./bin/hacktrackmmu-cli [flags]
```
**Flags:**
* `--config string`: Specify custom path to config file.
* `-v, --verbose`: Enable debug level structured logs (overrides `log_level` config value).

---

### Version Command
Print compilation details:
```bash
./bin/hacktrackmmu-cli version
```
Print in JSON format (e.g., for automated scripting or debugging):
```bash
./bin/hacktrackmmu-cli version --json
```

---

### Adding New Commands

To add a new subcommand to the CLI, use the `cobra-cli` binary:

```bash
# If not already installed:
go install github.com/spf13/cobra-cli@latest

# Add a command:
cobra-cli add <command-name>
```

Alternatively, you can manually create a new file in `cmd/<command-name>.go` following the pattern in `cmd/version.go`:

```go
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
	Use:   "mycmd",
	Short: "A brief description",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig() // Access the configuration
		fmt.Println("Running mycmd...")
	},
}

func init() {
	rootCmd.AddCommand(myCmd)
}
```
