package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xiaojian/goreadstat"
)

// CDISC Dataset-JSON 1.1 结构
type DatasetJSON struct {
	CreationDateTime string          `json:"datasetJSONCreationDateTime"`
	Version          string          `json:"datasetJSONVersion"`
	FileOID          string          `json:"fileOID,omitempty"`
	Originator       string          `json:"originator,omitempty"`
	SourceSystem     *SourceSystem   `json:"sourceSystem,omitempty"`
	StudyOID         string          `json:"studyOID,omitempty"`
	MetaDataRef      string          `json:"metaDataRef,omitempty"`
	ItemGroupOID     string          `json:"itemGroupOID,omitempty"`
	Records          int             `json:"records"`
	Name             string          `json:"name"`
	Label            string          `json:"label"`
	Columns          []ColumnDef     `json:"columns"`
	Rows             [][]interface{} `json:"rows,omitempty"`
}

type SourceSystem struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ColumnDef struct {
	ItemOID  string `json:"itemOID"`
	Name     string `json:"name"`
	Label    string `json:"label"`
	DataType string `json:"dataType"`
	Length   int    `json:"length,omitempty"`
}

func main() {
	baseDir := filepath.Dir(os.Args[0])
	if abs, err := filepath.Abs("."); err == nil {
		baseDir = abs
	}

	fmt.Println("=== goreadstat-v2 格式转换测试 ===")
	fmt.Printf("工作目录: %s\n\n", baseDir)

	// ──────────────────────────────────────────
	// 第一轮：XPT → SAS7BDAT, JSON
	// ──────────────────────────────────────────
	fmt.Println("━━━ 第一轮: XPT → SAS7BDAT / JSON ━━━")

	xptPath := filepath.Join(baseDir, "ec.xpt")
	df, meta, err := goreadstat.ReadXPORT(xptPath)
	mustOK("读取 ec.xpt", err)
	printSummary("ec.xpt", df, meta)

	sas7bdatPath := filepath.Join(baseDir, "ec_test.sas7bdat")
	err = goreadstat.WriteSAS(sas7bdatPath, df,
		goreadstat.WithFileLabel(meta.FileLabel),
		goreadstat.WithTableName(meta.TableName),
	)
	mustOK("XPT → ec_test.sas7bdat", err)

	jsonPath := filepath.Join(baseDir, "ec_test.json")
	err = writeDatasetJSON(jsonPath, df, meta)
	mustOK("XPT → ec_test.json", err)

	ndjsonPath := filepath.Join(baseDir, "ec_test.ndjson")
	err = writeDatasetNDJSON(ndjsonPath, df, meta)
	mustOK("XPT → ec_test.ndjson", err)

	// ──────────────────────────────────────────
	// 第二轮：JSON(NDJSON) → XPT, SAS7BDAT
	// ──────────────────────────────────────────
	fmt.Println("\n━━━ 第二轮: JSON → XPT / SAS7BDAT ━━━")

	srcNDJSON := filepath.Join(baseDir, "ec.ndjson")
	dfJ, metaJ, err := readDatasetNDJSON(srcNDJSON)
	mustOK("读取 ec.ndjson", err)
	printSummary("ec.ndjson", dfJ, metaJ)

	xptFromJSON := filepath.Join(baseDir, "ec_test_from_json.xpt")
	err = goreadstat.WriteXPORT(xptFromJSON, dfJ,
		goreadstat.WithFileLabel(metaJ.FileLabel),
		goreadstat.WithTableName(metaJ.TableName),
	)
	mustOK("JSON → ec_test_from_json.xpt", err)

	sasFromJSON := filepath.Join(baseDir, "ec_test_from_json.sas7bdat")
	err = goreadstat.WriteSAS(sasFromJSON, dfJ,
		goreadstat.WithFileLabel(metaJ.FileLabel),
		goreadstat.WithTableName(metaJ.TableName),
	)
	mustOK("JSON → ec_test_from_json.sas7bdat", err)

	// ──────────────────────────────────────────
	// 第三轮：SAS7BDAT → XPT, JSON
	// ──────────────────────────────────────────
	fmt.Println("\n━━━ 第三轮: SAS7BDAT → XPT / JSON ━━━")

	dfS, metaS, err := goreadstat.ReadSAS(sas7bdatPath)
	mustOK("读取 ec_test.sas7bdat", err)
	printSummary("ec_test.sas7bdat", dfS, metaS)

	xptFromSAS := filepath.Join(baseDir, "ec_test_from_sas.xpt")
	err = goreadstat.WriteXPORT(xptFromSAS, dfS,
		goreadstat.WithFileLabel(metaS.FileLabel),
		goreadstat.WithTableName(metaS.TableName),
	)
	mustOK("SAS7BDAT → ec_test_from_sas.xpt", err)

	jsonFromSAS := filepath.Join(baseDir, "ec_test_from_sas.ndjson")
	err = writeDatasetNDJSON(jsonFromSAS, dfS, metaS)
	mustOK("SAS7BDAT → ec_test_from_sas.ndjson", err)

	// ──────────────────────────────────────────
	// 验证：比对数据一致性
	// ──────────────────────────────────────────
	fmt.Println("\n━━━ 验证: 数据一致性比对 ━━━")

	verify("ec.xpt 原始", df, "ec_test.sas7bdat 转回", dfS)

	dfBack, _, err := goreadstat.ReadXPORT(xptFromJSON)
	mustOK("读取 ec_test_from_json.xpt", err)
	verify("ec.ndjson 原始", dfJ, "ec_test_from_json.xpt 转回", dfBack)

	dfBack2, _, err := goreadstat.ReadXPORT(xptFromSAS)
	mustOK("读取 ec_test_from_sas.xpt", err)
	verify("ec.xpt 原始", df, "ec_test_from_sas.xpt (XPT→SAS→XPT)", dfBack2)

	fmt.Println("\n=== 所有转换完成 ===")
	fmt.Println("\n生成的文件:")
	printGenerated(baseDir)
}

// ── JSON 读写 ──────────────────────────────

func writeDatasetJSON(path string, df *goreadstat.DataFrame, meta *goreadstat.Metadata) error {
	ds := buildDatasetJSON(df, meta)
	ds.Rows = df.Data

	data, err := json.MarshalIndent(ds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func writeDatasetNDJSON(path string, df *goreadstat.DataFrame, meta *goreadstat.Metadata) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	ds := buildDatasetJSON(df, meta)
	header, err := json.Marshal(ds)
	if err != nil {
		return err
	}
	w.Write(header)
	w.WriteByte('\n')

	for _, row := range df.Data {
		line, err := json.Marshal(row)
		if err != nil {
			return err
		}
		w.Write(line)
		w.WriteByte('\n')
	}
	return nil
}

func buildDatasetJSON(df *goreadstat.DataFrame, meta *goreadstat.Metadata) DatasetJSON {
	cols := make([]ColumnDef, len(df.Columns))
	for i, name := range df.Columns {
		col := ColumnDef{
			ItemOID:  fmt.Sprintf("IT.%s.%s", meta.TableName, name),
			Name:     name,
			DataType: "string",
		}

		if i < len(meta.Variables) {
			col.Label = meta.Variables[i].Label
			if meta.Variables[i].StorageWidth > 0 {
				col.Length = meta.Variables[i].StorageWidth
			}
		}

		if i < len(df.Types) && df.Types[i].IsNumeric() {
			isInt := isIntegerColumn(df.Data, i)
			if isInt {
				col.DataType = "integer"
			} else {
				col.DataType = "float"
			}
			col.Length = 0
		}

		cols[i] = col
	}

	name := meta.TableName
	if name == "" {
		name = "DATASET"
	}

	return DatasetJSON{
		CreationDateTime: meta.CreationTime.Format("2006-01-02T15:04:05"),
		Version:          "1.1.0",
		Records:          df.RowCount(),
		Name:             name,
		Label:            meta.FileLabel,
		Columns:          cols,
	}
}

func readDatasetNDJSON(path string) (*goreadstat.DataFrame, *goreadstat.Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	if !scanner.Scan() {
		return nil, nil, fmt.Errorf("empty ndjson file")
	}

	var ds DatasetJSON
	if err := json.Unmarshal(scanner.Bytes(), &ds); err != nil {
		return nil, nil, fmt.Errorf("parse header: %w", err)
	}

	columns := make([]string, len(ds.Columns))
	types := make([]goreadstat.ValueType, len(ds.Columns))
	variables := make([]goreadstat.Variable, len(ds.Columns))

	for i, col := range ds.Columns {
		columns[i] = col.Name
		switch col.DataType {
		case "integer", "float", "double", "decimal":
			types[i] = goreadstat.TypeDouble
		default:
			types[i] = goreadstat.TypeString
		}
		variables[i] = goreadstat.Variable{
			Index:        i,
			Name:         col.Name,
			Label:        col.Label,
			Type:         types[i],
			StorageWidth: col.Length,
		}
	}

	var rows [][]interface{}
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var rawRow []interface{}
		if err := json.Unmarshal(line, &rawRow); err != nil {
			return nil, nil, fmt.Errorf("parse row: %w", err)
		}
		row := make([]interface{}, len(rawRow))
		for i, val := range rawRow {
			if val == nil {
				row[i] = nil
				continue
			}
			if i < len(types) && types[i].IsNumeric() {
				switch v := val.(type) {
				case float64:
					row[i] = v
				case string:
					row[i] = v
				default:
					row[i] = val
				}
			} else {
				switch v := val.(type) {
				case string:
					row[i] = v
				case float64:
					row[i] = fmt.Sprintf("%g", v)
				default:
					row[i] = fmt.Sprintf("%v", v)
				}
			}
		}
		rows = append(rows, row)
	}

	df := &goreadstat.DataFrame{
		Columns: columns,
		Types:   types,
		Data:    rows,
	}

	meta := &goreadstat.Metadata{
		RowCount:    int64(len(rows)),
		ColumnCount: int64(len(columns)),
		TableName:   ds.Name,
		FileLabel:   ds.Label,
		Variables:   variables,
	}

	return df, meta, nil
}

// ── 辅助函数 ──────────────────────────────

func mustOK(action string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s 失败: %v\n", action, err)
		os.Exit(1)
	}
	fmt.Printf("  ✓ %s\n", action)
}

func printSummary(name string, df *goreadstat.DataFrame, meta *goreadstat.Metadata) {
	fmt.Printf("  源文件: %s → %d 行 × %d 列", name, df.RowCount(), df.ColumnCount())
	if meta.FileLabel != "" {
		fmt.Printf("  [%s]", meta.FileLabel)
	}
	fmt.Println()
}

func verify(nameA string, dfA *goreadstat.DataFrame, nameB string, dfB *goreadstat.DataFrame) {
	if dfA.RowCount() != dfB.RowCount() {
		fmt.Printf("  ✗ 行数不匹配: %s=%d, %s=%d\n", nameA, dfA.RowCount(), nameB, dfB.RowCount())
		return
	}
	if dfA.ColumnCount() != dfB.ColumnCount() {
		fmt.Printf("  ✗ 列数不匹配: %s=%d, %s=%d\n", nameA, dfA.ColumnCount(), nameB, dfB.ColumnCount())
		return
	}

	mismatches := 0
	for i := 0; i < dfA.RowCount() && i < 50; i++ {
		for j := 0; j < dfA.ColumnCount(); j++ {
			a := dfA.Data[i][j]
			b := dfB.Data[i][j]
			if !valuesEqual(a, b) {
				if mismatches < 3 {
					fmt.Printf("  ⚠ [%d][%s] 差异: %v ↔ %v\n", i, dfA.Columns[j], a, b)
				}
				mismatches++
			}
		}
	}

	if mismatches == 0 {
		fmt.Printf("  ✓ %s ↔ %s: %d行×%d列 数据一致\n",
			nameA, nameB, dfA.RowCount(), dfA.ColumnCount())
	} else {
		fmt.Printf("  ⚠ %s ↔ %s: 发现 %d 处差异 (前50行抽检)\n",
			nameA, nameB, mismatches)
	}
}

func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func isIntegerColumn(data [][]interface{}, colIdx int) bool {
	for _, row := range data {
		if colIdx >= len(row) || row[colIdx] == nil {
			continue
		}
		if v, ok := row[colIdx].(float64); ok {
			if v != float64(int64(v)) {
				return false
			}
		}
	}
	return true
}

func printGenerated(dir string) {
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), "_test") {
			info, _ := e.Info()
			fmt.Printf("  📄 %s  (%s)\n", e.Name(), humanSize(info.Size()))
		}
	}
}

func humanSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
}
