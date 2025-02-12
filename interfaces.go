package goparquet

import (
	"io"

	"github.com/sagia-inneractive/parquet-go/parquet"
)

// pageReader is an internal interface used only internally to read the pages
type pageReader interface {
	init(dDecoder, rDecoder getLevelDecoder, values getValueDecoderFn) error
	read(r io.Reader, ph *parquet.PageHeader, codec parquet.CompressionCodec) error

	readValues([]interface{}) (n int, dLevel *packedArray, rLevel *packedArray, err error)

	numValues() int32
}

// pageReader is an internal interface used only internally to read the pages
type pageWriter interface {
	init(schema SchemaWriter, col *Column, codec parquet.CompressionCodec) error

	write(w io.Writer) (int, int, error)
}

type newDataPageFunc func(useDict bool) pageWriter

type valuesDecoder interface {
	init(io.Reader) error
	// the error io.EOF with the less value is acceptable, any other error is not
	decodeValues([]interface{}) (int, error)
}

type dictValuesDecoder interface {
	valuesDecoder

	setValues([]interface{})
}

type valuesEncoder interface {
	init(io.Writer) error
	encodeValues([]interface{}) error

	io.Closer
}

type dictValuesEncoder interface {
	valuesEncoder

	getValues() []interface{}
}

// parquetColumn is to convert a store to a parquet.SchemaElement
type parquetColumn interface {
	parquetType() parquet.Type
	repetitionType() parquet.FieldRepetitionType
	params() *ColumnParameters
}

type typedColumnStore interface {
	parquetColumn
	reset(repetitionType parquet.FieldRepetitionType)
	// Min and Max in parquet byte
	maxValue() []byte
	minValue() []byte

	// Should extract the value, turn it into an array and check for min and max on all values in this
	getValues(v interface{}) ([]interface{}, error)
	sizeOf(v interface{}) int
	// the tricky append. this is a way of creating new "typed" array. the first interface is nil or an []T (T is the type,
	// not the interface) and value is from that type. the result should be always []T (array of that type)
	// exactly like the builtin append
	append(arrayIn interface{}, value interface{}) interface{}
}
