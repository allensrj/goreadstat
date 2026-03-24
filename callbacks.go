package goreadstat

/*
#include <readstat.h>
*/
import "C"

import (
	"sync"
	"time"
	"unsafe"
)

// --- Read-side context and registry ---

type parseContext struct {
	metadata      Metadata
	columns       []string
	types         []ValueType
	data          [][]interface{}
	varCount      int
	curRow        int
	selectedCols  map[string]bool
	colIndexMap   map[int]int // original index -> filtered index
}

var (
	readMu       sync.Mutex
	readNextID   uintptr
	readRegistry = make(map[uintptr]*parseContext)
)

func registerRead(ctx *parseContext) uintptr {
	readMu.Lock()
	defer readMu.Unlock()
	id := readNextID
	readNextID++
	readRegistry[id] = ctx
	return id
}

func lookupRead(ptr unsafe.Pointer) *parseContext {
	readMu.Lock()
	defer readMu.Unlock()
	id := *(*uintptr)(ptr)
	return readRegistry[id]
}

func unregisterRead(id uintptr) {
	readMu.Lock()
	defer readMu.Unlock()
	delete(readRegistry, id)
}

//export goMetadataHandler
func goMetadataHandler(metadata *C.readstat_metadata_t, ctx unsafe.Pointer) C.int {
	pctx := lookupRead(ctx)
	if pctx == nil {
		return C.READSTAT_HANDLER_ABORT
	}

	pctx.metadata.RowCount = int64(C.readstat_get_row_count(metadata))
	pctx.metadata.ColumnCount = int64(C.readstat_get_var_count(metadata))
	pctx.metadata.FormatVersion = int(C.readstat_get_file_format_version(metadata))
	pctx.metadata.Is64Bit = C.readstat_get_file_format_is_64bit(metadata) != 0
	pctx.metadata.Compression = Compression(C.readstat_get_compression(metadata))
	pctx.metadata.Endianness = Endian(C.readstat_get_endianness(metadata))

	ct := C.readstat_get_creation_time(metadata)
	if ct != 0 {
		pctx.metadata.CreationTime = time.Unix(int64(ct), 0)
	}
	mt := C.readstat_get_modified_time(metadata)
	if mt != 0 {
		pctx.metadata.ModifiedTime = time.Unix(int64(mt), 0)
	}

	if tn := C.readstat_get_table_name(metadata); tn != nil {
		pctx.metadata.TableName = C.GoString(tn)
	}
	if fl := C.readstat_get_file_label(metadata); fl != nil {
		pctx.metadata.FileLabel = C.GoString(fl)
	}
	if fe := C.readstat_get_file_encoding(metadata); fe != nil {
		pctx.metadata.FileEncoding = C.GoString(fe)
	}

	varCount := int(pctx.metadata.ColumnCount)
	pctx.varCount = varCount
	pctx.columns = make([]string, 0, varCount)
	pctx.types = make([]ValueType, 0, varCount)
	pctx.metadata.Variables = make([]Variable, 0, varCount)

	rowCount := pctx.metadata.RowCount
	if rowCount > 0 {
		pctx.data = make([][]interface{}, 0, rowCount)
	} else {
		pctx.data = make([][]interface{}, 0)
	}

	return C.READSTAT_HANDLER_OK
}

//export goVariableHandler
func goVariableHandler(index C.int, variable *C.readstat_variable_t, valLabels *C.char, ctx unsafe.Pointer) C.int {
	pctx := lookupRead(ctx)
	if pctx == nil {
		return C.READSTAT_HANDLER_ABORT
	}

	v := Variable{
		Index:        int(C.readstat_variable_get_index(variable)),
		StorageWidth: int(C.readstat_variable_get_storage_width(variable)),
		DisplayWidth: int(C.readstat_variable_get_display_width(variable)),
		Measure:      Measure(C.readstat_variable_get_measure(variable)),
		Alignment:    Alignment(C.readstat_variable_get_alignment(variable)),
	}

	cType := C.readstat_variable_get_type(variable)
	v.Type = ValueType(cType)

	if n := C.readstat_variable_get_name(variable); n != nil {
		v.Name = C.GoString(n)
	}
	if l := C.readstat_variable_get_label(variable); l != nil {
		v.Label = C.GoString(l)
	}
	if f := C.readstat_variable_get_format(variable); f != nil {
		v.Format = C.GoString(f)
	}
	if valLabels != nil {
		v.LabelSet = C.GoString(valLabels)
	}

	pctx.metadata.Variables = append(pctx.metadata.Variables, v)

	// If column filtering is enabled, only include selected columns
	if pctx.selectedCols != nil {
		if pctx.selectedCols[v.Name] {
			pctx.colIndexMap[v.Index] = len(pctx.columns)
			pctx.columns = append(pctx.columns, v.Name)
			pctx.types = append(pctx.types, v.Type)
		}
	} else {
		pctx.columns = append(pctx.columns, v.Name)
		pctx.types = append(pctx.types, v.Type)
	}

	return C.READSTAT_HANDLER_OK
}

//export goHandleValueDouble
func goHandleValueDouble(obsIndex C.int, varIndex C.int, value C.double, ctx unsafe.Pointer) C.int {
	pctx := lookupRead(ctx)
	if pctx == nil {
		return C.READSTAT_HANDLER_ABORT
	}

	oi := int(obsIndex)
	vi := int(varIndex)

	// Skip if column filtering is enabled and this column is not selected
	if pctx.selectedCols != nil {
		filteredIdx, ok := pctx.colIndexMap[vi]
		if !ok {
			return C.READSTAT_HANDLER_OK
		}
		vi = filteredIdx
	}

	colCount := len(pctx.columns)
	for len(pctx.data) <= oi {
		pctx.data = append(pctx.data, make([]interface{}, colCount))
	}

	if vi < len(pctx.data[oi]) {
		pctx.data[oi][vi] = float64(value)
	}

	return C.READSTAT_HANDLER_OK
}

//export goHandleValueString
func goHandleValueString(obsIndex C.int, varIndex C.int, value *C.char, ctx unsafe.Pointer) C.int {
	pctx := lookupRead(ctx)
	if pctx == nil {
		return C.READSTAT_HANDLER_ABORT
	}

	oi := int(obsIndex)
	vi := int(varIndex)

	if pctx.selectedCols != nil {
		filteredIdx, ok := pctx.colIndexMap[vi]
		if !ok {
			return C.READSTAT_HANDLER_OK
		}
		vi = filteredIdx
	}

	colCount := len(pctx.columns)
	for len(pctx.data) <= oi {
		pctx.data = append(pctx.data, make([]interface{}, colCount))
	}

	if vi < len(pctx.data[oi]) {
		if value != nil {
			pctx.data[oi][vi] = C.GoString(value)
		}
	}

	return C.READSTAT_HANDLER_OK
}

//export goHandleValueMissing
func goHandleValueMissing(obsIndex C.int, varIndex C.int, tag C.int, ctx unsafe.Pointer) C.int {
	pctx := lookupRead(ctx)
	if pctx == nil {
		return C.READSTAT_HANDLER_ABORT
	}

	oi := int(obsIndex)
	vi := int(varIndex)

	if pctx.selectedCols != nil {
		_, ok := pctx.colIndexMap[vi]
		if !ok {
			return C.READSTAT_HANDLER_OK
		}
	}

	colCount := len(pctx.columns)
	for len(pctx.data) <= oi {
		pctx.data = append(pctx.data, make([]interface{}, colCount))
	}

	return C.READSTAT_HANDLER_OK
}

//export goHandleValueLabel
func goHandleValueLabel(labelSet *C.char, valueType C.int, numKey C.double, strKey *C.char, label *C.char, ctx unsafe.Pointer) C.int {
	pctx := lookupRead(ctx)
	if pctx == nil {
		return C.READSTAT_HANDLER_ABORT
	}

	if pctx.metadata.ValueLabels == nil {
		pctx.metadata.ValueLabels = make(map[string][]ValueLabel)
	}

	setName := ""
	if labelSet != nil {
		setName = C.GoString(labelSet)
	}

	var vl ValueLabel
	if label != nil {
		vl.Label = C.GoString(label)
	}

	vType := ValueType(valueType)
	if vType == TypeString || vType == TypeStringRef {
		if strKey != nil {
			vl.Value = C.GoString(strKey)
		}
	} else {
		vl.Value = float64(numKey)
	}

	pctx.metadata.ValueLabels[setName] = append(pctx.metadata.ValueLabels[setName], vl)

	return C.READSTAT_HANDLER_OK
}

// --- Write-side context and registry ---

type writeContext struct {
	file interface{ Write([]byte) (int, error) }
	err  error
}

var (
	writeMu       sync.Mutex
	writeNextID   uintptr
	writeRegistry = make(map[uintptr]*writeContext)
)

func registerWrite(ctx *writeContext) uintptr {
	writeMu.Lock()
	defer writeMu.Unlock()
	id := writeNextID
	writeNextID++
	writeRegistry[id] = ctx
	return id
}

func lookupWrite(ptr unsafe.Pointer) *writeContext {
	writeMu.Lock()
	defer writeMu.Unlock()
	id := *(*uintptr)(ptr)
	return writeRegistry[id]
}

func unregisterWrite(id uintptr) {
	writeMu.Lock()
	defer writeMu.Unlock()
	delete(writeRegistry, id)
}

//export goWriteBytes
func goWriteBytes(data unsafe.Pointer, length C.size_t, ctx unsafe.Pointer) C.longlong {
	wctx := lookupWrite(ctx)
	if wctx == nil {
		return -1
	}

	buf := unsafe.Slice((*byte)(data), int(length))
	n, err := wctx.file.Write(buf)
	if err != nil {
		wctx.err = err
		return -1
	}
	return C.longlong(n)
}
