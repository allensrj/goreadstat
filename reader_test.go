package goreadstat

import (
	"os"
	"path/filepath"
	"testing"
)

var testDataDir = filepath.Join("test")

func testFileExists(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(testDataDir, name)
	if _, err := os.Stat(path); err != nil {
		t.Skipf("test file not found: %s", path)
	}
	return path
}

func TestReadXPORT(t *testing.T) {
	path := testFileExists(t, "ec.xpt")

	df, meta, err := ReadXPORT(path)
	if err != nil {
		t.Fatalf("ReadXPORT failed: %v", err)
	}

	if df.RowCount() == 0 {
		t.Fatal("expected rows, got 0")
	}
	if df.ColumnCount() == 0 {
		t.Fatal("expected columns, got 0")
	}

	t.Logf("XPORT: %d rows x %d columns", df.RowCount(), df.ColumnCount())
	t.Logf("  Table: %q, Label: %q", meta.TableName, meta.FileLabel)

	if len(meta.Variables) > 0 {
		v := meta.Variables[0]
		t.Logf("  First variable: name=%s label=%q", v.Name, v.Label)
	}
	if df.RowCount() > 0 && df.ColumnCount() > 3 {
		t.Logf("  Row 0, first 4 values: %v", df.Data[0][:4])
	}
}

func TestReadFile(t *testing.T) {
	path := testFileExists(t, "ec.xpt")

	df, _, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if df.RowCount() == 0 {
		t.Fatal("expected rows from ReadFile")
	}
	t.Logf("ReadFile auto-detect: %d rows x %d columns", df.RowCount(), df.ColumnCount())
}

func TestReadWithRowLimit(t *testing.T) {
	path := testFileExists(t, "ec.xpt")

	df, _, err := ReadXPORT(path, WithRowLimit(10))
	if err != nil {
		t.Fatalf("ReadXPORT with limit failed: %v", err)
	}

	if df.RowCount() > 10 {
		t.Fatalf("expected at most 10 rows, got %d", df.RowCount())
	}
	t.Logf("WithRowLimit(10): got %d rows", df.RowCount())
}

func TestColumnAccess(t *testing.T) {
	path := testFileExists(t, "ec.xpt")

	df, _, err := ReadXPORT(path, WithRowLimit(5))
	if err != nil {
		t.Fatalf("ReadXPORT failed: %v", err)
	}

	if len(df.Columns) == 0 {
		t.Fatal("no columns")
	}

	col, ok := df.Column(df.Columns[0])
	if !ok {
		t.Fatalf("Column(%q) returned false", df.Columns[0])
	}
	t.Logf("Column %q: %v", df.Columns[0], col)

	_, ok = df.Column("__nonexistent__")
	if ok {
		t.Fatal("expected false for nonexistent column")
	}
}

func TestWriteAndReadBack(t *testing.T) {
	df := &DataFrame{
		Columns: []string{"id", "name", "score"},
		Types:   []ValueType{TypeDouble, TypeString, TypeDouble},
		Data: [][]interface{}{
			{float64(1), "Alice", float64(95.5)},
			{float64(2), "Bob", float64(87.0)},
			{float64(3), "Charlie", nil},
		},
	}

	formats := []struct {
		name  string
		ext   string
		write func(string, *DataFrame, ...WriteOption) error
		read  func(string, ...Option) (*DataFrame, *Metadata, error)
	}{
		{"DTA", ".dta", WriteDTA, ReadDTA},
		{"SAV", ".sav", WriteSAV, ReadSAV},
		{"SAS", ".sas7bdat", WriteSAS, ReadSAS},
		{"XPORT", ".xpt", WriteXPORT, ReadXPORT},
	}

	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "test"+f.ext)

			err := f.write(tmpFile, df, WithFileLabel("test data"))
			if err != nil {
				t.Fatalf("Write%s failed: %v", f.name, err)
			}

			readDF, meta, err := f.read(tmpFile)
			if err != nil {
				t.Fatalf("Read%s failed: %v", f.name, err)
			}

			if readDF.ColumnCount() != df.ColumnCount() {
				t.Errorf("column count: got %d, want %d", readDF.ColumnCount(), df.ColumnCount())
			}
			if readDF.RowCount() != df.RowCount() {
				t.Errorf("row count: got %d, want %d", readDF.RowCount(), df.RowCount())
			}

			t.Logf("%s round-trip: %d rows x %d cols, label=%q",
				f.name, readDF.RowCount(), readDF.ColumnCount(), meta.FileLabel)

			for i, row := range readDF.Data {
				t.Logf("  Row %d: %v", i, row)
			}
		})
	}
}

func TestReadWithColumns(t *testing.T) {
	path := testFileExists(t, "ec.xpt")

	df, _, err := ReadXPORT(path, WithColumns([]string{"STUDYID", "USUBJID"}), WithRowLimit(5))
	if err != nil {
		t.Fatalf("ReadXPORT with columns failed: %v", err)
	}

	if df.ColumnCount() != 2 {
		t.Fatalf("expected 2 columns, got %d", df.ColumnCount())
	}

	if df.Columns[0] != "STUDYID" || df.Columns[1] != "USUBJID" {
		t.Fatalf("unexpected columns: %v", df.Columns)
	}

	t.Logf("WithColumns: %d rows x %d cols", df.RowCount(), df.ColumnCount())
	for i, row := range df.Data {
		t.Logf("  Row %d: %v", i, row)
	}
}

func TestReadNonExistentFile(t *testing.T) {
	_, _, err := ReadSAS("nonexistent.sas7bdat")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	t.Logf("Got expected error: %v", err)
}

func TestReadWithInvalidColumns(t *testing.T) {
	path := testFileExists(t, "ec.xpt")

	df, _, err := ReadXPORT(path, WithColumns([]string{"NONEXISTENT"}), WithRowLimit(5))
	if err != nil {
		t.Fatalf("ReadXPORT failed: %v", err)
	}

	if df.ColumnCount() != 0 {
		t.Fatalf("expected 0 columns for invalid column names, got %d", df.ColumnCount())
	}
	t.Logf("Invalid columns correctly returned empty dataframe")
}

func TestReadEmptyColumns(t *testing.T) {
	path := testFileExists(t, "ec.xpt")

	df, _, err := ReadXPORT(path, WithColumns([]string{}), WithRowLimit(5))
	if err != nil {
		t.Fatalf("ReadXPORT failed: %v", err)
	}

	if df.ColumnCount() == 0 {
		t.Fatal("expected all columns when empty slice provided")
	}
	t.Logf("Empty columns list correctly returned all columns: %d", df.ColumnCount())
}
