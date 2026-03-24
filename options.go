package goreadstat

// readOptions holds configuration for reading files.
type readOptions struct {
	rowLimit  int64
	rowOffset int64
	encoding  string
	columns   []string
}

// Option configures read operations.
type Option func(*readOptions)

// WithRowLimit limits the number of rows to read.
func WithRowLimit(n int64) Option {
	return func(o *readOptions) { o.rowLimit = n }
}

// WithRowOffset skips the first n rows.
func WithRowOffset(n int64) Option {
	return func(o *readOptions) { o.rowOffset = n }
}

// WithEncoding specifies the character encoding (e.g., "UTF-8", "GBK").
func WithEncoding(enc string) Option {
	return func(o *readOptions) { o.encoding = enc }
}

// WithColumns specifies which columns to read. If empty, all columns are read.
func WithColumns(cols []string) Option {
	return func(o *readOptions) { o.columns = cols }
}

type writeOptions struct {
	fileLabel   string
	tableName   string
	compression Compression
	version     int
}

// WriteOption configures write operations.
type WriteOption func(*writeOptions)

// WithFileLabel sets the file label/description.
func WithFileLabel(label string) WriteOption {
	return func(o *writeOptions) { o.fileLabel = label }
}

// WithTableName sets the table/dataset name.
func WithTableName(name string) WriteOption {
	return func(o *writeOptions) { o.tableName = name }
}

// WithCompression sets the compression method.
func WithCompression(c Compression) WriteOption {
	return func(o *writeOptions) { o.compression = c }
}

// WithVersion sets the file format version.
func WithVersion(v int) WriteOption {
	return func(o *writeOptions) { o.version = v }
}
