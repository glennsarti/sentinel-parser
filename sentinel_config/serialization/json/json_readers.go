package json

import (
	"encoding/json"
	"fmt"

	"github.com/glennsarti/sentinel-parser/position"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
)

func readValue[T any](data json.RawMessage) (*T, error) {
	var raw T
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

func readValueList[T any](data json.RawMessage, reader func(rawJsonObject) (*T, error)) (*[]*T, error) {
	rawList := make([]json.RawMessage, 0)
	if err := json.Unmarshal(data, &rawList); err != nil {
		return nil, err
	}

	list := make([]*T, len(rawList))
	for idx, rawItem := range rawList {
		raw, err := readJsonBytes(rawItem)
		if err != nil {
			return nil, err
		} else if raw != nil {
			actual, err := reader(*raw)
			if err != nil {
				return nil, err
			}
			list[idx] = actual
		}
	}

	return &list, nil
}

func readValueStringMapPointer[T any](data json.RawMessage, reader func(rawJsonObject) (*T, error)) (*map[string]*T, error) {
	rawList := make(map[string]json.RawMessage, 0)
	if err := json.Unmarshal(data, &rawList); err != nil {
		return nil, err
	}

	list := make(map[string]*T, len(rawList))
	for idx, rawItem := range rawList {
		raw, err := readJsonBytes(rawItem)
		if err != nil {
			return nil, err
		} else if raw != nil {
			actual, err := reader(*raw)
			if err != nil {
				return nil, err
			}
			list[idx] = actual
		}
	}

	return &list, nil
}

func readValueStringMap[T any](data json.RawMessage, reader func(rawJsonObject) (T, error)) (*map[string]T, error) {
	rawList := make(map[string]json.RawMessage, 0)
	if err := json.Unmarshal(data, &rawList); err != nil {
		return nil, err
	}

	list := make(map[string]T, len(rawList))
	for idx, rawItem := range rawList {
		raw, err := readJsonBytes(rawItem)
		if err != nil {
			return nil, err
		} else if raw != nil {
			actual, err := reader(*raw)
			if err != nil {
				return nil, err
			}
			list[idx] = actual
		}
	}

	return &list, nil
}

func isJsonNull(data json.RawMessage) bool {
	return len(data) == 4 &&
		data[0] == 110 &&
		data[1] == 117 &&
		data[2] == 108 &&
		data[3] == 108
}

func readJsonBytes(data []byte) (*rawJsonObject, error) {
	var raw rawJsonObject
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

func (jo *rawJsonObject) readProperty(name string) (*rawJsonObject, error) {
	if data, ok := jo.Properties[name]; !ok {
		return nil, nil
	} else if isJsonNull(data) {
		return nil, nil
	} else {
		raw := rawJsonObject{}
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
		return &raw, nil
	}
}

func (jo *rawJsonObject) readPropertyBytes(name string) (*json.RawMessage, error) {
	if data, ok := jo.Properties[name]; !ok {
		return nil, nil
	} else if isJsonNull(data) {
		return nil, nil
	} else {
		return &data, nil
	}
}

func (jc *jsonCoder) readSourceRange(data *json.RawMessage) (*position.SourceRange, error) {
	if data == nil {
		return nil, nil
	}
	obj := position.SourceRange{}
	if err := json.Unmarshal(*data, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (jc *jsonCoder) readDynamicValue(data rawJsonObject) (*ast.DynamicValue, error) {
	if data.AstType != "DynamicValue" {
		return nil, fmt.Errorf("expected '_t' property in 'Feature' object to be 'DynamicValue' but got %q", data.AstType)
	}

	if raw, ok := data.Properties["value"]; !ok {
		return nil, nil
	} else {
		val := ast.DynamicValue{}
		if err := val.UnmarshalJSON(raw); err != nil {
			return nil, err
		} else {
			return &val, nil
		}
	}
}
