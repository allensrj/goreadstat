package goreadstat

import "fmt"

type ReadStatError struct {
	Code    int
	Message string
}

func (e *ReadStatError) Error() string {
	return fmt.Sprintf("readstat error %d: %s", e.Code, e.Message)
}

var errorMessages = map[int]string{
	0:  "OK",
	1:  "open error",
	2:  "read error",
	3:  "malloc error",
	4:  "user abort",
	5:  "parse error",
	6:  "unsupported compression",
	7:  "unsupported charset",
	8:  "column count mismatch",
	9:  "row count mismatch",
	10: "row width mismatch",
	11: "bad format string",
	12: "value type mismatch",
	13: "write error",
	14: "writer not initialized",
	15: "seek error",
	16: "convert error",
	17: "convert bad string",
	18: "convert short string",
	19: "convert long string",
	20: "numeric value out of range",
	21: "tagged value out of range",
	22: "string value too long",
	23: "tagged values not supported",
	24: "unsupported file format version",
	25: "name begins with illegal character",
	26: "name contains illegal character",
	27: "name is reserved word",
	28: "name is too long",
	29: "bad timestamp string",
	30: "bad frequency weight",
	31: "too many missing value definitions",
	32: "note is too long",
	33: "string refs not supported",
	34: "string ref is required",
	35: "row is too wide for page",
	36: "too few columns",
	37: "too many columns",
	38: "name is zero length",
	39: "bad timestamp value",
	40: "bad MR string",
}

func newReadStatError(code int) error {
	if code == 0 {
		return nil
	}
	msg, ok := errorMessages[code]
	if !ok {
		msg = "unknown error"
	}
	return &ReadStatError{Code: code, Message: msg}
}
