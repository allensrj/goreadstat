# goreadstat

Go package for reading and writing SAS, SPSS, and Stata statistical data files. Powered by the [ReadStat](https://github.com/WizardMac/ReadStat) C library via cgo.

[![Go Reference](https://pkg.go.dev/badge/github.com/allensrj/goreadstat.svg)](https://pkg.go.dev/github.com/allensrj/goreadstat)
[![Go Report Card](https://goreportcard.com/badge/github.com/allensrj/goreadstat)](https://goreportcard.com/report/github.com/allensrj/goreadstat)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **Zero dependencies** - Direct C binding, no Python or external processes
- **High performance** - Native C library performance
- **Rich metadata** - Variable labels, formats, value labels, file attributes
- **Flexible reading** - Row limits, offsets, column selection, encoding support
- **Complete write support** - Create files in all major formats

## Supported Formats

| Format | Extension | Read | Write |
|--------|-----------|------|-------|
| SAS binary | `.sas7bdat` | Yes | Yes |
| SAS transport | `.xpt` | Yes | Yes |
| SAS catalog | `.sas7bcat` | Yes | — |
| SPSS binary | `.sav` / `.zsav` | Yes | Yes |
| SPSS portable | `.por` | Yes | Yes |
| Stata | `.dta` | Yes | Yes |

## Prerequisites

Before using this library, ensure you have:

- **Go 1.21+**
- **C compiler** (gcc or clang)
- **make**
- **zlib** (pre-installed on macOS/Linux)
- **iconv** (pre-installed on macOS; part of glibc on Linux)

### Platform-specific setup

**macOS:**
```bash
# Usually no additional setup needed
xcode-select --install  # If you don't have command line tools
```

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install build-essential zlib1g-dev
```

**Other Linux:**
```bash
# Install gcc, make, and zlib-devel through your package manager
```

## Installation

### Step 1: Get the package

```bash
go get github.com/allensrj/goreadstat@latest
```

### Step 2: Build the C library

**Important:** You must build the ReadStat C library before using this package.

```bash
# Navigate to the package directory
cd $(go list -f '{{.Dir}}' github.com/allensrj/goreadstat)

# Build the C library (one-time setup)
make

# Verify installation
go test -v
```

### Step 3: Use in your project

```go
package main

import (
    "fmt"
    "log"
    "github.com/allensrj/goreadstat"
)

func main() {
    df, meta, err := goreadstat.ReadFile("data.sas7bdat")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%d rows × %d columns\n", df.RowCount(), df.ColumnCount())
}
```

Then run:
```bash
go mod tidy
go run main.go
```

## Quick Start

### Reading Files

```go
package main

import (
    "fmt"
    "github.com/xiaojian/goreadstat"
)

func main() {
    // Auto-detect format by extension
    df, meta, err := goreadstat.ReadFile("data.sas7bdat")
    if err != nil {
        panic(err)
    }

    fmt.Printf("%d rows x %d columns\n", df.RowCount(), df.ColumnCount())
    fmt.Printf("Label: %s\n", meta.FileLabel)

    // Access column data
    col, _ := df.Column("AGE")
    fmt.Println(col)
}
```

### Format-Specific Readers

```go
df, meta, err := goreadstat.ReadSAS("file.sas7bdat")
df, meta, err := goreadstat.ReadXPORT("file.xpt")
df, meta, err := goreadstat.ReadSAV("file.sav")
df, meta, err := goreadstat.ReadPOR("file.por")
df, meta, err := goreadstat.ReadDTA("file.dta")
```

### Read Options

```go
// Limit rows
df, meta, err := goreadstat.ReadSAS("big.sas7bdat",
    goreadstat.WithRowLimit(100),
    goreadstat.WithRowOffset(50),
)

// Select specific columns
df, meta, err := goreadstat.ReadSAS("data.sas7bdat",
    goreadstat.WithColumns([]string{"id", "name", "age"}),
)

// Specify encoding
df, meta, err := goreadstat.ReadDTA("old.dta",
    goreadstat.WithEncoding("GBK"),
)

// Combine options
df, meta, err := goreadstat.ReadSAV("data.sav",
    goreadstat.WithColumns([]string{"var1", "var2"}),
    goreadstat.WithRowLimit(1000),
)
```

### Writing Files

```go
df := &goreadstat.DataFrame{
    Columns: []string{"id", "name", "score"},
    Types:   []goreadstat.ValueType{
        goreadstat.TypeDouble,
        goreadstat.TypeString,
        goreadstat.TypeDouble,
    },
    Data: [][]interface{}{
        {float64(1), "Alice", float64(95.5)},
        {float64(2), "Bob",   float64(87.0)},
        {float64(3), "Carol", nil},  // nil = missing value
    },
}

goreadstat.WriteSAS("output.sas7bdat", df)
goreadstat.WriteXPORT("output.xpt", df)
goreadstat.WriteSAV("output.sav", df, goreadstat.WithFileLabel("My Data"))
goreadstat.WriteDTA("output.dta", df)
goreadstat.WritePOR("output.por", df)
```

## Data Types

- Numeric columns → `float64`
- String columns → `string`
- Missing values → `nil`

## Metadata

The `Metadata` struct provides rich file information:

```go
df, meta, _ := goreadstat.ReadFile("data.sas7bdat")

meta.RowCount       // number of rows
meta.ColumnCount    // number of columns
meta.FileLabel      // file label
meta.FileEncoding   // character encoding
meta.CreationTime   // creation timestamp
meta.Variables      // detailed variable info (name, label, format, type)
meta.ValueLabels    // value label sets (common in SPSS/Stata)
```

## Project Structure

```
├── go.mod               # Go module
├── Makefile             # Builds ReadStat static library
├── readstat/            # ReadStat C library source (MIT)
├── readstat_helpers.c/h # C callback wrappers
├── cgo.go               # cgo configuration
├── callbacks.go         # Go↔C callback bridge (core)
├── types.go             # Type definitions
├── errors.go            # Error mapping
├── options.go           # Functional options
├── reader.go            # Read API
├── writer.go            # Write API
├── reader_test.go       # Tests
├── test/                # Test data & conversion script
└── examples/            # CLI example
```

## How It Works

This library uses cgo to bind directly to the ReadStat C library. The key technical challenge — passing Go callbacks to C — is solved using a **global registry pattern**:

1. A Go `parseContext` is registered in a thread-safe map, receiving an integer ID
2. The ID is passed to ReadStat as `void* user_ctx`
3. C callback wrappers extract data from ReadStat's C types and call Go-exported functions
4. The Go functions look up the context by ID and populate the DataFrame

This achieves native C performance with zero runtime dependencies (no Python, no subprocess).

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

The bundled ReadStat C library is Copyright Evan Miller and contributors, also under the MIT License.
