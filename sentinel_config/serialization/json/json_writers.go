package json

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/glennsarti/sentinel-parser/position"
	"github.com/glennsarti/sentinel-parser/sentinel_config/ast"
)

func (jc *jsonCoder) writeHeader(typeName string, writer io.Writer) error {
	_, err := fmt.Fprintf(writer, `{"_t":"%s","_p":{`, typeName)
	return err
}

func (jc *jsonCoder) writeFooter(writer io.Writer) error {
	_, err := fmt.Fprintf(writer, `}}`)
	return err
}

func (jc *jsonCoder) writeDynamicValue(dv *ast.DynamicValue, writer io.Writer) error {
	if dv == nil {
		_, err := fmt.Fprintf(writer, "null")
		return err
	}

	if err := jc.writeHeader("DynamicValue", writer); err != nil {
		return nil
	}
	if _, err := fmt.Fprint(writer, `"value":`); err != nil {
		return err
	}

	buf, err := dv.MarshalJSON()
	if err != nil {
		return err
	}

	if _, err = writer.Write(buf); err != nil {
		return err
	}
	return jc.writeFooter(writer)
}

func (jc *jsonCoder) writeLiteralString(value string, writer io.Writer) error {
	if v, err := json.Marshal(value); err != nil {
		return err
	} else {
		_, err := writer.Write(v)
		return err
	}
}

func (jc *jsonCoder) writeLiteralStringList(values []string, writer io.Writer) error {
	if _, err := fmt.Fprint(writer, `[`); err != nil {
		return nil
	}
	for idx, value := range values {
		if idx > 0 {
			if _, err := fmt.Fprint(writer, `,`); err != nil {
				return err
			}
		}
		if err := jc.writeLiteralString(value, writer); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(writer, `]`)
	return err
}

func (jc *jsonCoder) writeSourceRange(src *position.SourceRange, writer io.Writer) error {
	if src == nil {
		_, err := fmt.Fprintf(writer, "null")
		return err
	}

	if v, err := json.Marshal(src); err != nil {
		return err
	} else {
		_, err := writer.Write(v)
		return err
	}
}

func writeList[T ast.Node](obj []T, jc *jsonCoder, writer io.Writer) error {
	if _, err := fmt.Fprint(writer, `[`); err != nil {
		return err
	}
	for idx, item := range obj {
		if idx > 0 {
			if _, err := fmt.Fprint(writer, `,`); err != nil {
				return err
			}
		}
		if err := jc.write(item, writer); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(writer, `]`)
	return err
}

func writeStringMap[T ast.Node](obj map[string]T, jc *jsonCoder, writer io.Writer) error {
	if _, err := fmt.Fprint(writer, `{`); err != nil {
		return err
	}

	// Use a stable order for the keys
	// _Technically_ order shouldn't matter, but the number of globals/prarams/mocks/... etc.
	// probably won't be large (100+) so performance-wise this _shouldn't_ be an issue.
	keys := make([]string, len(obj))
	i := 0
	for key := range obj {
		keys[i] = key
		i += 1
	}
	sort.Strings(keys)

	// Write out each key/value
	for idx, key := range keys {
		if idx > 0 {
			if _, err := fmt.Fprint(writer, `,`); err != nil {
				return err
			}
		}
		if err := jc.writeLiteralString(key, writer); err != nil {
			return err
		}
		if _, err := fmt.Fprint(writer, `:`); err != nil {
			return err
		}
		if err := jc.write(obj[key], writer); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(writer, `}`)
	return err
}
