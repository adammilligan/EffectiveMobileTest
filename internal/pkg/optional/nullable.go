package optional

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// NullableString supports 3 states:
// - IsSet=false: field absent in JSON
// - IsSet=true, Value=nil: field is null
// - IsSet=true, Value!=nil: field has value
type NullableString struct {
	IsSet bool
	Value *string
}

func (n *NullableString) UnmarshalJSON(data []byte) error {
	n.IsSet = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		n.Value = nil
		return nil
	}

	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	n.Value = &v

	return nil
}

type NullableInt struct {
	IsSet bool
	Value *int
}

func (n *NullableInt) UnmarshalJSON(data []byte) error {
	n.IsSet = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		n.Value = nil
		return nil
	}

	var v int
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	n.Value = &v

	return nil
}

type NullableMonth struct {
	IsSet bool
	Value *string
}

func (n *NullableMonth) UnmarshalJSON(data []byte) error {
	n.IsSet = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		n.Value = nil
		return nil
	}

	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if v == "" {
		return fmt.Errorf("month is empty")
	}

	n.Value = &v

	return nil
}

