package goreadstat

/*
#include <readstat.h>
#include <stdlib.h>
#include "readstat_helpers.h"
*/
import "C"

import (
	"fmt"
	"math"
	"os"
	"unsafe"
)

func writeFile(path string, df *DataFrame, format FileFormat, opts ...WriteOption) error {
	var wo writeOptions
	for _, o := range opts {
		o(&wo)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	wctx := &writeContext{file: f}
	id := registerWrite(wctx)
	defer unregisterWrite(id)

	writer := C.readstat_writer_init()
	if writer == nil {
		return fmt.Errorf("failed to initialize writer")
	}
	defer C.readstat_writer_free(writer)

	C.readstat_set_data_writer(writer, C.readstat_data_writer(C.c_data_writer))

	if wo.fileLabel != "" {
		cl := C.CString(wo.fileLabel)
		defer C.free(unsafe.Pointer(cl))
		C.readstat_writer_set_file_label(writer, cl)
	}
	if wo.tableName != "" {
		ct := C.CString(wo.tableName)
		defer C.free(unsafe.Pointer(ct))
		C.readstat_writer_set_table_name(writer, ct)
	}
	if wo.compression != CompressNone {
		C.readstat_writer_set_compression(writer, C.readstat_compress_t(wo.compression))
	}
	if wo.version > 0 {
		C.readstat_writer_set_file_format_version(writer, C.uint8_t(wo.version))
	}

	cVars := make([]*C.readstat_variable_t, len(df.Columns))
	for i, colName := range df.Columns {
		cName := C.CString(colName)
		defer C.free(unsafe.Pointer(cName))

		var cType C.readstat_type_t
		var storageWidth C.size_t

		if i < len(df.Types) && df.Types[i].IsNumeric() {
			cType = C.READSTAT_TYPE_DOUBLE
			storageWidth = 8
		} else {
			cType = C.READSTAT_TYPE_STRING
			storageWidth = C.size_t(maxStringWidth(df.Data, i))
		}

		cVars[i] = C.readstat_add_variable(writer, cName, cType, storageWidth)
		if cVars[i] == nil {
			return fmt.Errorf("failed to add variable: %s", colName)
		}
	}

	userCtx := unsafe.Pointer(&id)
	rowCount := C.long(len(df.Data))

	var rc C.readstat_error_t
	switch format {
	case FormatDTA:
		rc = C.readstat_begin_writing_dta(writer, userCtx, rowCount)
	case FormatSAV:
		rc = C.readstat_begin_writing_sav(writer, userCtx, rowCount)
	case FormatPOR:
		rc = C.readstat_begin_writing_por(writer, userCtx, rowCount)
	case FormatXPORT:
		rc = C.readstat_begin_writing_xport(writer, userCtx, rowCount)
	case FormatSAS7BDAT:
		rc = C.readstat_begin_writing_sas7bdat(writer, userCtx, rowCount)
	default:
		return fmt.Errorf("unsupported write format: %d", format)
	}
	if rc != C.READSTAT_OK {
		return newReadStatError(int(rc))
	}

	for _, row := range df.Data {
		rc = C.readstat_begin_row(writer)
		if rc != C.READSTAT_OK {
			return newReadStatError(int(rc))
		}

		for colIdx, val := range row {
			if colIdx >= len(cVars) {
				break
			}
			variable := cVars[colIdx]

			if val == nil {
				rc = C.readstat_insert_missing_value(writer, variable)
			} else {
				switch v := val.(type) {
				case float64:
					if math.IsNaN(v) || math.IsInf(v, 0) {
						rc = C.readstat_insert_missing_value(writer, variable)
					} else {
						rc = C.readstat_insert_double_value(writer, variable, C.double(v))
					}
				case int:
					rc = C.readstat_insert_double_value(writer, variable, C.double(v))
				case int64:
					rc = C.readstat_insert_double_value(writer, variable, C.double(v))
				case int32:
					rc = C.readstat_insert_double_value(writer, variable, C.double(v))
				case string:
					cs := C.CString(v)
					rc = C.readstat_insert_string_value(writer, variable, cs)
					C.free(unsafe.Pointer(cs))
				default:
					rc = C.readstat_insert_missing_value(writer, variable)
				}
			}

			if rc != C.READSTAT_OK {
				return newReadStatError(int(rc))
			}
		}

		rc = C.readstat_end_row(writer)
		if rc != C.READSTAT_OK {
			return newReadStatError(int(rc))
		}
	}

	rc = C.readstat_end_writing(writer)
	if rc != C.READSTAT_OK {
		return newReadStatError(int(rc))
	}

	if wctx.err != nil {
		return fmt.Errorf("write error: %w", wctx.err)
	}

	return nil
}

// WriteSAS writes a DataFrame to a SAS binary file (.sas7bdat).
func WriteSAS(path string, df *DataFrame, opts ...WriteOption) error {
	return writeFile(path, df, FormatSAS7BDAT, opts...)
}

// WriteDTA writes a DataFrame to a Stata file (.dta).
func WriteDTA(path string, df *DataFrame, opts ...WriteOption) error {
	return writeFile(path, df, FormatDTA, opts...)
}

// WriteSAV writes a DataFrame to an SPSS binary file (.sav).
func WriteSAV(path string, df *DataFrame, opts ...WriteOption) error {
	return writeFile(path, df, FormatSAV, opts...)
}

// WritePOR writes a DataFrame to an SPSS portable file (.por).
func WritePOR(path string, df *DataFrame, opts ...WriteOption) error {
	return writeFile(path, df, FormatPOR, opts...)
}

// WriteXPORT writes a DataFrame to a SAS transport file (.xpt).
func WriteXPORT(path string, df *DataFrame, opts ...WriteOption) error {
	return writeFile(path, df, FormatXPORT, opts...)
}

func maxStringWidth(data [][]interface{}, colIdx int) int {
	maxLen := 8
	for _, row := range data {
		if colIdx < len(row) {
			if s, ok := row[colIdx].(string); ok {
				if len(s) > maxLen {
					maxLen = len(s)
				}
			}
		}
	}
	return maxLen
}
