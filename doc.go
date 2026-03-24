// Package goreadstat provides reading and writing support for SAS, SPSS, and Stata statistical data files.
//
// It uses the ReadStat C library via cgo for high-performance native file format support.
//
// # Supported Formats
//
// Read and write:
//   - SAS: .sas7bdat (binary), .xpt (transport)
//   - SPSS: .sav (binary), .zsav (compressed), .por (portable)
//   - Stata: .dta
//
// Read only:
//   - SAS: .sas7bcat (catalog)
//
// # Basic Usage
//
// Reading a file:
//
//	df, meta, err := goreadstat.ReadFile("data.sas7bdat")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("%d rows × %d columns\n", df.RowCount(), df.ColumnCount())
//
// Writing a file:
//
//	df := &goreadstat.DataFrame{
//	    Columns: []string{"id", "name"},
//	    Types:   []goreadstat.ValueType{goreadstat.TypeDouble, goreadstat.TypeString},
//	    Data: [][]interface{}{
//	        {1.0, "Alice"},
//	        {2.0, "Bob"},
//	    },
//	}
//	err := goreadstat.WriteSAS("output.sas7bdat", df)
//
// # Options
//
// Read operations support various options:
//
//	df, meta, err := goreadstat.ReadSAS("data.sas7bdat",
//	    goreadstat.WithColumns([]string{"id", "name"}),
//	    goreadstat.WithRowLimit(1000),
//	    goreadstat.WithRowOffset(100),
//	)
//
// Write operations also support options:
//
//	err := goreadstat.WriteSAV("output.sav", df,
//	    goreadstat.WithFileLabel("Survey Data"),
//	    goreadstat.WithCompression(goreadstat.CompressRows),
//	)
package goreadstat
