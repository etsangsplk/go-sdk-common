package ldvalue

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonMarshalUnmarshal(t *testing.T) {
	items := []struct {
		value Value
		json  string
	}{
		{Null(), "null"},
		{Bool(true), "true"},
		{Bool(false), "false"},
		{Int(1), "1"},
		{Float64(1), "1"},
		{Float64(2.5), "2.5"},
		{String("x"), `"x"`},
		{ArrayBuild().Add(Bool(true)).Add(String("x")).Build(), `[true,"x"]`},
		{ObjectBuild().Set("a", Bool(true)).Build(), `{"a":true}`},
	}
	for _, item := range items {
		t.Run(fmt.Sprintf("type %s, json %v", item.value.Type(), item.json), func(t *testing.T) {
			j, err := json.Marshal(item.value)
			assert.NoError(t, err)
			assert.Equal(t, item.json, string(j))

			assert.Equal(t, item.json, item.value.String())
			assert.Equal(t, item.json, item.value.JSONString())
			assert.Equal(t, json.RawMessage(item.json), item.value.AsRaw())

			var v Value
			err = json.Unmarshal([]byte(item.json), &v)
			assert.NoError(t, err)
			assert.Equal(t, item.value, v)
		})
	}
}

func TestUnmarshalErrorConditions(t *testing.T) {
	var v Value
	assert.Error(t, json.Unmarshal(nil, &v))
	assert.Error(t, json.Unmarshal([]byte{}, &v))
	assert.Error(t, json.Unmarshal([]byte("what"), &v))
}

func TestMarshalWithUnexpectedError(t *testing.T) {
	// This can only happen if there's some custom type within an unsafe complex value
	// that has its own marshalling method that fails.
	sliceWithWeirdValue := []interface{}{valueThatRefusesToBeMarshalled{}}
	v := UnsafeUseArbitraryValue(sliceWithWeirdValue)
	_, err := json.Marshal(v)
	assert.Error(t, err)
	jsonString := v.JSONString()
	assert.Equal(t, "", jsonString)
	raw := v.AsRaw()
	assert.Nil(t, raw)

	// Calling CopyArbitraryValue on an unknown type causes it to be marshalled; if that
	// fails, we're supposed to just use null
	v1 := CopyArbitraryValue(valueThatRefusesToBeMarshalled{})
	assert.Equal(t, Null(), v1)
}

type valueThatRefusesToBeMarshalled struct{}

func (v valueThatRefusesToBeMarshalled) MarshalJSON() ([]byte, error) {
	return nil, errors.New("sorry")
}
