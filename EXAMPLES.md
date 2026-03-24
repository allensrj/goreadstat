# goreadstat Examples

## Basic Reading

### Read any format (auto-detect)

```go
df, meta, err := goreadstat.ReadFile("data.sas7bdat")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%d rows × %d columns\n", df.RowCount(), df.ColumnCount())
```

### Read specific format

```go
df, meta, err := goreadstat.ReadSAS("file.sas7bdat")
df, meta, err := goreadstat.ReadXPORT("file.xpt")
df, meta, err := goreadstat.ReadSAV("file.sav")
df, meta, err := goreadstat.ReadDTA("file.dta")
```

## Advanced Reading

### Select specific columns

```go
df, meta, err := goreadstat.ReadSAS("large.sas7bdat",
    goreadstat.WithColumns([]string{"id", "name", "score"}),
)
```

### Read subset of rows

```go
// Skip first 100 rows, read next 50
df, meta, err := goreadstat.ReadSAV("data.sav",
    goreadstat.WithRowOffset(100),
    goreadstat.WithRowLimit(50),
)
```

### Handle encoding

```go
df, meta, err := goreadstat.ReadDTA("chinese.dta",
    goreadstat.WithEncoding("GBK"),
)
```

## Working with Data

### Access columns

```go
df, _, _ := goreadstat.ReadFile("data.sas7bdat")

// Get column by name
ages, ok := df.Column("age")
if ok {
    for i, val := range ages {
        if val != nil {
            fmt.Printf("Row %d: %v\n", i, val)
        }
    }
}
```

### Iterate rows

```go
for i, row := range df.Data {
    fmt.Printf("Row %d: ", i)
    for j, val := range row {
        if val == nil {
            fmt.Printf("%s=<missing> ", df.Columns[j])
        } else {
            fmt.Printf("%s=%v ", df.Columns[j], val)
        }
    }
    fmt.Println()
}
```

## Metadata

### File information

```go
_, meta, _ := goreadstat.ReadFile("data.sav")

fmt.Println("File Label:", meta.FileLabel)
fmt.Println("Encoding:", meta.FileEncoding)
fmt.Println("Created:", meta.CreationTime)
fmt.Println("Rows:", meta.RowCount)
```

### Variable information

```go
for _, v := range meta.Variables {
    fmt.Printf("Variable: %s\n", v.Name)
    fmt.Printf("  Label: %s\n", v.Label)
    fmt.Printf("  Format: %s\n", v.Format)
    fmt.Printf("  Type: %s\n", v.Type)
}
```

### Value labels

```go
for setName, labels := range meta.ValueLabels {
    fmt.Printf("Label set: %s\n", setName)
    for _, vl := range labels {
        fmt.Printf("  %v = %s\n", vl.Value, vl.Label)
    }
}
```

## Writing Files

### Basic write

```go
df := &goreadstat.DataFrame{
    Columns: []string{"id", "name", "score"},
    Types: []goreadstat.ValueType{
        goreadstat.TypeDouble,
        goreadstat.TypeString,
        goreadstat.TypeDouble,
    },
    Data: [][]interface{}{
        {1.0, "Alice", 95.5},
        {2.0, "Bob", 87.0},
        {3.0, "Carol", nil}, // missing value
    },
}

err := goreadstat.WriteSAS("output.sas7bdat", df)
```

### Write with options

```go
err := goreadstat.WriteSAV("output.sav", df,
    goreadstat.WithFileLabel("Survey Results 2024"),
    goreadstat.WithTableName("survey"),
    goreadstat.WithCompression(goreadstat.CompressRows),
)
```

## Error Handling

```go
df, meta, err := goreadstat.ReadFile("data.sas7bdat")
if err != nil {
    if os.IsNotExist(err) {
        log.Fatal("File not found")
    }
    log.Fatalf("Read error: %v", err)
}

if df.RowCount() == 0 {
    log.Println("Warning: empty dataset")
}
```

## Performance Tips

1. Use column selection to reduce memory usage:
```go
df, _, _ := goreadstat.ReadSAS("huge.sas7bdat",
    goreadstat.WithColumns([]string{"id", "date"}),
)
```

2. Process data in chunks:
```go
chunkSize := 10000
for offset := 0; ; offset += chunkSize {
    df, _, err := goreadstat.ReadSAS("large.sas7bdat",
        goreadstat.WithRowOffset(int64(offset)),
        goreadstat.WithRowLimit(int64(chunkSize)),
    )
    if err != nil || df.RowCount() == 0 {
        break
    }
    // Process chunk
}
```
