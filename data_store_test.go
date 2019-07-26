package go_parquet

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/fraugster/parquet-go/parquet"
)

func newIntStore(rep parquet.FieldRepetitionType) columnStore {
	d := newStore(&int32Store{})
	d.reset(rep)
	return d
}

func newIntStore2() columnStore {
	d := newStore(&int32Store{})
	return d
}

func TestOneColumn(t *testing.T) {
	row := rowStore{}
	require.NoError(t, row.addColumn("DocID", newIntStore2(), parquet.FieldRepetitionType_REQUIRED))

	data := []map[string]interface{}{
		{"DocID": int32(10)},
		{"DocID": int32(20)},
	}

	for i := range data {
		require.NoError(t, row.add(data[i]))
	}
	d, err := row.findDataColumn("DocID")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int32(10), int32(20)}, d.dictionary().assemble())
	assert.Equal(t, []int32{0, 0}, d.definitionLevels())
	assert.Equal(t, []int32{0, 0}, d.repetitionLevels())
}

func TestOneColumnOptional(t *testing.T) {
	row := rowStore{}
	require.NoError(t, row.addColumn("DocID", newIntStore2(), parquet.FieldRepetitionType_OPTIONAL))

	data := []map[string]interface{}{
		{"DocID": int32(10)},
		{},
	}

	for i := range data {
		require.NoError(t, row.add(data[i]))
	}
	d, err := row.findDataColumn("DocID")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int32(10), nil}, d.dictionary().assemble())
	assert.Equal(t, []int32{1, 0}, d.definitionLevels())
	assert.Equal(t, []int32{0, 0}, d.repetitionLevels())
}

func TestComplexPart1(t *testing.T) {
	row := &rowStore{}
	require.NoError(t, row.addGroup("Name", parquet.FieldRepetitionType_REPEATED))
	require.NoError(t, row.addGroup("Name.Language", parquet.FieldRepetitionType_REPEATED))
	require.NoError(t, row.addColumn("Name.Language.Code", newIntStore2(), parquet.FieldRepetitionType_REQUIRED))
	require.NoError(t, row.addColumn("Name.Language.Country", newIntStore2(), parquet.FieldRepetitionType_OPTIONAL))
	require.NoError(t, row.addColumn("Name.URL", newIntStore2(), parquet.FieldRepetitionType_OPTIONAL))

	data := []map[string]interface{}{
		{
			"Name": []map[string]interface{}{
				{
					"Language": []map[string]interface{}{
						{
							"Code":    int32(1),
							"Country": int32(100),
						},
						{
							"Code": int32(2),
						},
					},
					"URL": int32(10),
				},
				{
					"URL": int32(11),
				},
				{
					"Language": []map[string]interface{}{
						{
							"Code":    int32(3),
							"Country": int32(101),
						},
					},
				},
			},
		},
	}

	for i := range data {
		require.NoError(t, row.add(data[i]))
	}

	d, err := row.findDataColumn("Name.Language.Code")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int32(1), int32(2), nil, int32(3)}, d.dictionary().assemble())
	assert.Equal(t, []int32{2, 2, 1, 2}, d.definitionLevels())
	assert.Equal(t, []int32{0, 2, 1, 1}, d.repetitionLevels())

	d, err = row.findDataColumn("Name.Language.Country")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int32(100), nil, nil, int32(101)}, d.dictionary().assemble())
	assert.Equal(t, []int32{3, 2, 1, 3}, d.definitionLevels())
	assert.Equal(t, []int32{0, 2, 1, 1}, d.repetitionLevels())

	d, err = row.findDataColumn("Name.URL")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int32(10), int32(11), nil}, d.dictionary().assemble())
	assert.Equal(t, []int32{2, 2, 1}, d.definitionLevels())
	assert.Equal(t, []int32{0, 1, 1}, d.repetitionLevels())

}

func TestComplexPart2(t *testing.T) {
	row := &rowStore{}
	require.NoError(t, row.addGroup("Links", parquet.FieldRepetitionType_OPTIONAL))
	require.NoError(t, row.addColumn("Links.Backward", newIntStore2(), parquet.FieldRepetitionType_REPEATED))
	require.NoError(t, row.addColumn("Links.Forward", newIntStore2(), parquet.FieldRepetitionType_REPEATED))

	data := []map[string]interface{}{
		{
			"Links": map[string]interface{}{
				"Forward": []int32{20, 40, 60},
			},
		},
		{
			"Links": map[string]interface{}{
				"Backward": []int32{10, 30},
				"Forward":  []int32{80},
			},
		},
	}

	for i := range data {
		require.NoError(t, row.add(data[i]))
	}

	d, err := row.findDataColumn("Links.Forward")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int32(20), int32(40), int32(60), int32(80)}, d.dictionary().assemble())
	assert.Equal(t, []int32{2, 2, 2, 2}, d.definitionLevels())
	assert.Equal(t, []int32{0, 1, 1, 0}, d.repetitionLevels())

	d, err = row.findDataColumn("Links.Backward")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{nil, int32(10), int32(30)}, d.dictionary().assemble())
	assert.Equal(t, []int32{1, 2, 2}, d.definitionLevels())
	assert.Equal(t, []int32{0, 0, 1}, d.repetitionLevels())
}

func TestComplex(t *testing.T) {
	// Based on this picture https://i.stack.imgur.com/raOFu.png from this doc https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/36632.pdf
	row := &rowStore{}
	require.NoError(t, row.addColumn("DocId", newIntStore2(), parquet.FieldRepetitionType_REQUIRED))
	require.NoError(t, row.addGroup("Links", parquet.FieldRepetitionType_OPTIONAL))
	require.NoError(t, row.addColumn("Links.Backward", newIntStore2(), parquet.FieldRepetitionType_REPEATED))
	require.NoError(t, row.addColumn("Links.Forward", newIntStore2(), parquet.FieldRepetitionType_REPEATED))
	require.NoError(t, row.addGroup("Name", parquet.FieldRepetitionType_REPEATED))
	require.NoError(t, row.addGroup("Name.Language", parquet.FieldRepetitionType_REPEATED))
	require.NoError(t, row.addColumn("Name.Language.Code", newIntStore2(), parquet.FieldRepetitionType_REQUIRED))
	require.NoError(t, row.addColumn("Name.Language.Country", newIntStore2(), parquet.FieldRepetitionType_OPTIONAL))
	require.NoError(t, row.addColumn("Name.URL", newIntStore2(), parquet.FieldRepetitionType_OPTIONAL))

	data := []map[string]interface{}{
		{
			"DocId": int32(10),
			"Links": map[string]interface{}{
				"Forward": []int32{20, 40, 60},
			},
			"Name": []map[string]interface{}{
				{
					"Language": []map[string]interface{}{
						{
							"Code":    int32(1),
							"Country": int32(100),
						},
						{
							"Code": int32(2),
						},
					},
					"URL": int32(10),
				},
				{
					"URL": int32(11),
				},
				{
					"Language": []map[string]interface{}{
						{
							"Code":    int32(3),
							"Country": int32(101),
						},
					},
				},
			},
		},
		{
			"DocId": int32(20),
			"Links": map[string]interface{}{
				"Backward": []int32{10, 30},
				"Forward":  []int32{80},
			},
			"Name": []map[string]interface{}{
				{
					"URL": int32(12),
				},
			},
		},
	}

	for i := range data {
		require.NoError(t, row.add(data[i]))
	}

	d, err := row.findDataColumn("DocId")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int32(10), int32(20)}, d.dictionary().assemble())
	assert.Equal(t, []int32{0, 0}, d.definitionLevels())
	assert.Equal(t, []int32{0, 0}, d.repetitionLevels())

	d, err = row.findDataColumn("Name.URL")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int32(10), int32(11), nil, int32(12)}, d.dictionary().assemble())
	assert.Equal(t, []int32{2, 2, 1, 2}, d.definitionLevels())
	assert.Equal(t, []int32{0, 1, 1, 0}, d.repetitionLevels())

	d, err = row.findDataColumn("Links.Forward")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int32(20), int32(40), int32(60), int32(80)}, d.dictionary().assemble())
	assert.Equal(t, []int32{2, 2, 2, 2}, d.definitionLevels())
	assert.Equal(t, []int32{0, 1, 1, 0}, d.repetitionLevels())

	d, err = row.findDataColumn("Links.Backward")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{nil, int32(10), int32(30)}, d.dictionary().assemble())
	assert.Equal(t, []int32{1, 2, 2}, d.definitionLevels())
	assert.Equal(t, []int32{0, 0, 1}, d.repetitionLevels())

	d, err = row.findDataColumn("Name.Language.Country")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int32(100), nil, nil, int32(101), nil}, d.dictionary().assemble())
	assert.Equal(t, []int32{3, 2, 1, 3, 1}, d.definitionLevels())
	assert.Equal(t, []int32{0, 2, 1, 1, 0}, d.repetitionLevels())

	d, err = row.findDataColumn("Name.Language.Code")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{int32(1), int32(2), nil, int32(3), nil}, d.dictionary().assemble())
	assert.Equal(t, []int32{2, 2, 1, 2, 1}, d.definitionLevels())
	assert.Equal(t, []int32{0, 2, 1, 1, 0}, d.repetitionLevels())

}

func TestTwitterBlog(t *testing.T) {
	// Sample from here https://blog.twitter.com/engineering/en_us/a/2013/dremel-made-simple-with-parquet.html
	row := &rowStore{}
	require.NoError(t, row.addGroup("level1", parquet.FieldRepetitionType_REPEATED))
	require.NoError(t, row.addColumn("level1.level2", newIntStore2(), parquet.FieldRepetitionType_REPEATED))

	data := []map[string]interface{}{
		{
			"level1": []map[string]interface{}{
				{"level2": []int32{1, 2, 3}},
				{"level2": []int32{4, 5, 6, 7}},
			},
		},
		{
			"level1": []map[string]interface{}{
				{"level2": []int32{8}},
				{"level2": []int32{9, 10}},
			},
		},
	}

	for i := range data {
		require.NoError(t, row.add(data[i]))
	}

	d, err := row.findDataColumn("level1.level2")
	require.NoError(t, err)
	var expected []interface{}
	for i := 1; i < 11; i++ {
		expected = append(expected, int32(i))
	}
	assert.Equal(t, expected, d.dictionary().assemble())
	assert.Equal(t, []int32{0, 2, 2, 1, 2, 2, 2, 0, 1, 2}, d.repetitionLevels())

}
