package goreadstat

import "time"

// ValueType represents the data type of a variable.
type ValueType int

const (
	TypeString ValueType = iota
	TypeInt8
	TypeInt16
	TypeInt32
	TypeFloat
	TypeDouble
	TypeStringRef
)

func (vt ValueType) String() string {
	switch vt {
	case TypeString:
		return "string"
	case TypeInt8:
		return "int8"
	case TypeInt16:
		return "int16"
	case TypeInt32:
		return "int32"
	case TypeFloat:
		return "float"
	case TypeDouble:
		return "double"
	case TypeStringRef:
		return "string_ref"
	default:
		return "unknown"
	}
}

func (vt ValueType) IsNumeric() bool {
	return vt != TypeString && vt != TypeStringRef
}

// Compression represents the compression method used in a file.
type Compression int

const (
	CompressNone   Compression = 0
	CompressRows   Compression = 1
	CompressBinary Compression = 2
)

// Endian represents byte order.
type Endian int

const (
	EndianNone   Endian = 0
	EndianLittle Endian = 1
	EndianBig    Endian = 2
)

// Measure represents the measurement level of a variable (SPSS).
type Measure int

const (
	MeasureUnknown Measure = 0
	MeasureNominal Measure = 1
	MeasureOrdinal Measure = 2
	MeasureScale   Measure = 3
)

// Alignment represents text alignment for display.
type Alignment int

const (
	AlignUnknown Alignment = 0
	AlignLeft    Alignment = 1
	AlignCenter  Alignment = 2
	AlignRight   Alignment = 3
)

// Variable contains metadata about a single variable/column.
type Variable struct {
	Index        int
	Name         string
	Label        string
	Format       string
	Type         ValueType
	StorageWidth int
	DisplayWidth int
	Measure      Measure
	Alignment    Alignment
	LabelSet     string
}

// ValueLabel represents a value-to-label mapping.
type ValueLabel struct {
	Value interface{} // float64 or string
	Label string
}

// Metadata contains file-level information and variable metadata.
type Metadata struct {
	RowCount      int64
	ColumnCount   int64
	CreationTime  time.Time
	ModifiedTime  time.Time
	FormatVersion int
	Is64Bit       bool
	Compression   Compression
	Endianness    Endian
	TableName     string
	FileLabel     string
	FileEncoding  string
	Variables     []Variable
	ValueLabels   map[string][]ValueLabel
}

// DataFrame holds tabular data returned by read functions.
// Each element of Data is a row; within a row, values are ordered by column.
// Numeric values are float64; strings are string; missing values are nil.
type DataFrame struct {
	Columns []string
	Types   []ValueType
	Data    [][]interface{}
}

func (df *DataFrame) RowCount() int {
	return len(df.Data)
}

func (df *DataFrame) ColumnCount() int {
	return len(df.Columns)
}

func (df *DataFrame) Column(name string) ([]interface{}, bool) {
	idx := -1
	for i, col := range df.Columns {
		if col == name {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, false
	}
	result := make([]interface{}, len(df.Data))
	for i, row := range df.Data {
		if idx < len(row) {
			result[i] = row[idx]
		}
	}
	return result, true
}

type FileFormat int

const (
	FormatDTA      FileFormat = iota // Stata
	FormatSAV                       // SPSS
	FormatPOR                       // SPSS portable
	FormatSAS7BDAT                  // SAS binary
	FormatSAS7BCAT                  // SAS catalog
	FormatXPORT                     // SAS transport
)
