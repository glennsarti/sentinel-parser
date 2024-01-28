package json

import (
	"encoding/json"
	"fmt"

	"github.com/glennsarti/sentinel-parser/position"
)

func readValue[T any](data json.RawMessage) (*T, error) {
	var raw T
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

func readValueList[T any](data json.RawMessage, reader func(rawJsonObject) (T, error)) (*[]T, error) {
	rawList := make([]json.RawMessage, 0)
	if err := json.Unmarshal(data, &rawList); err != nil {
		return nil, err
	}

	list := make([]T, len(rawList))
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

func readValuePtrList[T any](data json.RawMessage, reader func(rawJsonObject) (*T, error)) (*[]*T, error) {
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

func (jc *jsonCoder) readSourceRangePtr(data *json.RawMessage) (*position.SourceRange, error) {
	if data == nil {
		return nil, nil
	}

	raw := jsonSourceRange{}
	if err := json.Unmarshal(*data, &raw); err != nil {
		return nil, err
	}

	return &position.SourceRange{
		End: position.SourcePos{
			Byte:   raw.End.Byte,
			Column: raw.End.Column,
			Line:   raw.End.Line,
		},
		Filename: raw.Filename,
		Start: position.SourcePos{
			Byte:   raw.Start.Byte,
			Column: raw.Start.Column,
			Line:   raw.Start.Line,
		},
	}, nil
}

func (jc *jsonCoder) readSourceRange(data *json.RawMessage) (position.SourceRange, error) {
	if v, err := jc.readSourceRangePtr(data); err != nil {
		return position.SourceRange{}, err
	} else if v == nil {
		return position.SourceRange{}, fmt.Errorf("could not read a SourceRange")
	} else {
		return *v, nil
	}
}
