package goreadstat

/*
#include <readstat.h>
#include <stdlib.h>
#include "readstat_helpers.h"
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

func readFile(path string, format FileFormat, opts ...Option) (*DataFrame, *Metadata, error) {
	var ro readOptions
	for _, o := range opts {
		o(&ro)
	}

	if _, err := os.Stat(path); err != nil {
		return nil, nil, fmt.Errorf("file not found: %s", path)
	}

	parser := C.readstat_parser_init()
	if parser == nil {
		return nil, nil, fmt.Errorf("failed to initialize parser")
	}
	defer C.readstat_parser_free(parser)

	if err := C.setup_read_handlers(parser); err != C.READSTAT_OK {
		return nil, nil, newReadStatError(int(err))
	}

	if ro.rowLimit > 0 {
		C.readstat_set_row_limit(parser, C.long(ro.rowLimit))
	}
	if ro.rowOffset > 0 {
		C.readstat_set_row_offset(parser, C.long(ro.rowOffset))
	}
	if ro.encoding != "" {
		enc := C.CString(ro.encoding)
		defer C.free(unsafe.Pointer(enc))
		C.readstat_set_file_character_encoding(parser, enc)
	}

	pctx := &parseContext{}
	if len(ro.columns) > 0 {
		pctx.selectedCols = make(map[string]bool)
		for _, col := range ro.columns {
			pctx.selectedCols[col] = true
		}
		pctx.colIndexMap = make(map[int]int)
	}
	id := registerRead(pctx)
	defer unregisterRead(id)

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	userCtx := unsafe.Pointer(&id)

	var rc C.readstat_error_t
	switch format {
	case FormatSAS7BDAT:
		rc = C.readstat_parse_sas7bdat(parser, cpath, userCtx)
	case FormatXPORT:
		rc = C.readstat_parse_xport(parser, cpath, userCtx)
	case FormatSAV:
		rc = C.readstat_parse_sav(parser, cpath, userCtx)
	case FormatPOR:
		rc = C.readstat_parse_por(parser, cpath, userCtx)
	case FormatDTA:
		rc = C.readstat_parse_dta(parser, cpath, userCtx)
	case FormatSAS7BCAT:
		rc = C.readstat_parse_sas7bcat(parser, cpath, userCtx)
	default:
		return nil, nil, fmt.Errorf("unsupported format: %d", format)
	}

	if rc != C.READSTAT_OK {
		return nil, nil, newReadStatError(int(rc))
	}

	df := &DataFrame{
		Columns: pctx.columns,
		Types:   pctx.types,
		Data:    pctx.data,
	}
	meta := &pctx.metadata

	return df, meta, nil
}

// ReadSAS reads a SAS binary file (.sas7bdat).
// Returns a DataFrame with the data, Metadata with file information, and any error encountered.
func ReadSAS(path string, opts ...Option) (*DataFrame, *Metadata, error) {
	return readFile(path, FormatSAS7BDAT, opts...)
}

// ReadXPORT reads a SAS transport file (.xpt).
func ReadXPORT(path string, opts ...Option) (*DataFrame, *Metadata, error) {
	return readFile(path, FormatXPORT, opts...)
}

// ReadSAV reads an SPSS binary file (.sav or .zsav).
func ReadSAV(path string, opts ...Option) (*DataFrame, *Metadata, error) {
	return readFile(path, FormatSAV, opts...)
}

// ReadPOR reads an SPSS portable file (.por).
func ReadPOR(path string, opts ...Option) (*DataFrame, *Metadata, error) {
	return readFile(path, FormatPOR, opts...)
}

// ReadDTA reads a Stata file (.dta).
func ReadDTA(path string, opts ...Option) (*DataFrame, *Metadata, error) {
	return readFile(path, FormatDTA, opts...)
}

// ReadSAS7BCAT reads a SAS catalog file (.sas7bcat).
// Returns only Metadata as catalog files don't contain tabular data.
func ReadSAS7BCAT(path string, opts ...Option) (*Metadata, error) {
	_, meta, err := readFile(path, FormatSAS7BCAT, opts...)
	return meta, err
}

// ReadFile automatically detects the file format by extension and reads it.
// Supported extensions: .sas7bdat, .xpt, .sav, .zsav, .por, .dta
func ReadFile(path string, opts ...Option) (*DataFrame, *Metadata, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".sas7bdat":
		return ReadSAS(path, opts...)
	case ".xpt", ".xport":
		return ReadXPORT(path, opts...)
	case ".sav", ".zsav":
		return ReadSAV(path, opts...)
	case ".por":
		return ReadPOR(path, opts...)
	case ".dta":
		return ReadDTA(path, opts...)
	default:
		return nil, nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}
