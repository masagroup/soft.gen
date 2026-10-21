package props

import (
	"fmt"
	"io"

	"go.yaml.in/yaml/v3"
)

// yamlFormat parses a YAML file holding a flat mapping. Every scalar value is
// kept as its YAML text (bools and numbers as written, strings as is, block
// scalars with their lines); null is "". A nested mapping or sequence value is
// an error naming the key; an empty document is an empty mapping.
type yamlFormat struct{}

func init() { register(yamlFormat{}) }

func (yamlFormat) name() string         { return "yaml" }
func (yamlFormat) extensions() []string { return []string{".yaml", ".yml"} }

func (yamlFormat) parse(r io.Reader) (map[string]string, error) {
	var doc yaml.Node
	if err := yaml.NewDecoder(r).Decode(&doc); err != nil {
		if err == io.EOF {
			return map[string]string{}, nil
		}
		return nil, err
	}
	root := resolveYAMLAlias(&doc)
	if root.Kind == yaml.DocumentNode {
		if len(root.Content) == 0 {
			return map[string]string{}, nil
		}
		root = resolveYAMLAlias(root.Content[0])
	}
	if root.Kind == yaml.ScalarNode && root.ShortTag() == "!!null" {
		return map[string]string{}, nil
	}
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("line %d: expected a mapping of properties", root.Line)
	}
	values := make(map[string]string, len(root.Content)/2)
	for i := 0; i+1 < len(root.Content); i += 2 {
		keyNode, valueNode := resolveYAMLAlias(root.Content[i]), resolveYAMLAlias(root.Content[i+1])
		if keyNode.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("line %d: property keys must be scalars", keyNode.Line)
		}
		key := keyNode.Value
		switch {
		case valueNode.Kind != yaml.ScalarNode:
			return nil, fmt.Errorf("line %d: property %q: value must be a scalar, not a mapping or a sequence", valueNode.Line, key)
		case valueNode.ShortTag() == "!!null":
			values[key] = ""
		default:
			values[key] = valueNode.Value
		}
	}
	return values, nil
}

// resolveYAMLAlias follows alias nodes to the anchored node.
func resolveYAMLAlias(n *yaml.Node) *yaml.Node {
	for n.Kind == yaml.AliasNode && n.Alias != nil {
		n = n.Alias
	}
	return n
}
