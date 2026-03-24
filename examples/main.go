package main

import (
	"fmt"
	"os"

	"github.com/xiaojian/goreadstat"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <file>")
		fmt.Println("Supported: .sas7bdat .xpt .sav .por .dta")
		os.Exit(1)
	}

	path := os.Args[1]

	df, meta, err := goreadstat.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File:     %s\n", path)
	fmt.Printf("Rows:     %d\n", df.RowCount())
	fmt.Printf("Columns:  %d\n", df.ColumnCount())
	fmt.Println()

	if meta.TableName != "" {
		fmt.Printf("Table:    %s\n", meta.TableName)
	}
	if meta.FileLabel != "" {
		fmt.Printf("Label:    %s\n", meta.FileLabel)
	}
	if meta.FileEncoding != "" {
		fmt.Printf("Encoding: %s\n", meta.FileEncoding)
	}
	if !meta.CreationTime.IsZero() {
		fmt.Printf("Created:  %s\n", meta.CreationTime.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("Version:  %d (64-bit: %v)\n", meta.FormatVersion, meta.Is64Bit)
	fmt.Println()

	fmt.Println("Variables:")
	for _, v := range meta.Variables {
		label := ""
		if v.Label != "" {
			label = " — " + v.Label
		}
		fmt.Printf("  [%d] %s (%s, format=%s)%s\n", v.Index, v.Name, v.Type, v.Format, label)
	}

	if len(meta.ValueLabels) > 0 {
		fmt.Printf("\nValue Labels: %d sets\n", len(meta.ValueLabels))
		for name, labels := range meta.ValueLabels {
			fmt.Printf("  %s: %d labels\n", name, len(labels))
		}
	}

	limit := 5
	if df.RowCount() < limit {
		limit = df.RowCount()
	}
	fmt.Printf("\nFirst %d rows:\n", limit)
	for i := 0; i < limit; i++ {
		fmt.Printf("  [%d] %v\n", i, df.Data[i])
	}
}
