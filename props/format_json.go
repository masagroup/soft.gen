package props

import (
	"encoding/json"
	"fmt"
	"io"
)

// jsonFormat parses a JSON file holding a flat object. Strings are kept as is,
// numbers as written (no float artefacts), booleans as "true"/"false" and null
// as "". A nested object or array value is an error naming the key.
type jsonFormat struct{}

func init() { register(jsonFormat{}) }

func (jsonFormat) name() string         { return "json" }
func (jsonFormat) extensions() []string { return []string{".json"} }

func (jsonFormat) parse(r io.Reader) (map[string]string, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	var raw map[string]any
	if err := dec.Decode(&raw); err != nil {
		if err == io.EOF {
			return map[string]string{}, nil
		}
		return nil, err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("unexpected content after the property object")
	}
	values := make(map[string]string, len(raw))
	for key, v := range raw {
		switch v := v.(type) {
		case nil:
			values[key] = ""
		case string:
			values[key] = v
		case json.Number:
			values[key] = v.String()
		case bool:
			if v {
				values[key] = "true"
			} else {
				values[key] = "false"
			}
		default:
			return nil, fmt.Errorf("property %q: value must be a scalar, not an object or an array", key)
		}
	}
	return values, nil
}
