package ast

import (
	"encoding/json"

	"github.com/zclconf/go-cty/cty"
	ctyJson "github.com/zclconf/go-cty/cty/json"
)

type DynamicValue struct {
	value cty.Value
}

func ConvertCtyValue(value *cty.Value) *DynamicValue {
	if value == nil {
		return nil
	}
	obj := DynamicValue{value: *value}
	return &obj
}

func (dv DynamicValue) GoString() string {
	return dv.value.GoString()
}

func (dv DynamicValue) Clone() DynamicValue {
	// This is a bit naughty but it'll do for now
	// cty.Value is not mutatable
	return DynamicValue{
		value: dv.value,
	}
}

func (dv *DynamicValue) MarshalJSON() ([]byte, error) {
	sv := ctyJson.SimpleJSONValue{Value: dv.value}
	return json.Marshal(sv)
}

func (dv *DynamicValue) UnmarshalJSON(data []byte) error {

	sv := &ctyJson.SimpleJSONValue{}
	if err := sv.UnmarshalJSON(data); err != nil {
		return err
	}

	dv.value = sv.Value
	return nil
}

func DynamicValueComparer(x, y DynamicValue) bool {
	return x.value.GoString() == y.value.GoString()
}

func DynamicValuePtrComparer(x, y *DynamicValue) bool {
	if x == nil && y == nil {
		return true
	}
	if x != nil && y != nil {
		return x.value.GoString() == y.value.GoString()
	}
	return false
}
