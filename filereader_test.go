package goparquet

import (
	"bytes"
	"io"
	"math/rand"
	"testing"

	"github.com/sagia-inneractive/parquet-go/parquetschema"
	"github.com/stretchr/testify/require"
)

func buildTestStream(t *testing.T) []byte {
	schema, err := parquetschema.ParseSchemaDefinition(`message msg {
  required int64 a;
  required int64 b;
  optional group x {
    required int64 c;
    required int64 d;
  }
  required group y {
     required int64 e;
  }
}
`)
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	pw := NewFileWriter(buf, WithSchemaDefinition(schema))
	for i := 0; i < 10000; i++ {
		data := map[string]interface{}{
			"a": rand.Int63(),
			"b": rand.Int63(),
			"x": map[string]interface{}{
				"c": rand.Int63(),
				"d": rand.Int63(),
			},
			"y": map[string]interface{}{
				"e": rand.Int63(),
			},
		}
		require.NoError(t, pw.AddData(data))
		if i%100 == 0 {
			require.NoError(t, pw.FlushRowGroup())
		}
	}
	require.NoError(t, pw.Close())
	return buf.Bytes()
}

func TestByteReaderSelected(t *testing.T) {
	r := buildTestStream(t)
	pr, err := NewFileReader(bytes.NewReader(r), "a")
	require.NoError(t, err)

	for {
		data, err := pr.NextRow()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Equal(t, 2, len(data))
		_, ok := data["a"]
		require.True(t, ok)
		y, ok := data["y"]
		require.True(t, ok)
		require.Empty(t, y)
	}
}

func TestByteReaderSelectedInner(t *testing.T) {
	r := buildTestStream(t)
	pr, err := NewFileReader(bytes.NewReader(r), "x.c")
	require.NoError(t, err)

	for {
		data, err := pr.NextRow()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Equal(t, 2, len(data))
		x, ok := data["x"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, 1, len(x))
		y, ok := data["y"]
		require.True(t, ok)
		require.Empty(t, y)
	}
}

func TestByteReaderSelectedInnerFull(t *testing.T) {
	r := buildTestStream(t)
	pr, err := NewFileReader(bytes.NewReader(r), "x")
	require.NoError(t, err)

	for {
		data, err := pr.NextRow()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Equal(t, 2, len(data))
		x, ok := data["x"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, 2, len(x))
		y, ok := data["y"]
		require.True(t, ok)
		require.Empty(t, y)
	}
}
