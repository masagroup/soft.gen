package model

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"sort"
	"strconv"
)

// soft.go does not model the EMF generics (EGenericType, ETypeParameter): the
// xml decoder reports their elements as unknown features and drops them, with
// the type an element only has through its eGenericType (eInvoke(...,
// arguments EEList), eContents() EEList). v1 ran on EMF, where eType derives
// from eGenericType. stripGenerics restores that before decoding.

// stripGenerics returns the .ecore document data without its generic
// elements: an eGenericType with an eClassifier gives its parent the eType
// attribute when the parent has none (the EMF derivation of eType), then the
// eGenericType and eTypeParameters sub trees are removed. The rest of the
// document is left byte for byte.
func stripGenerics(data []byte) ([]byte, error) {
	stripped, _, err := stripGenericsWithArguments(data)
	return stripped, err
}

// typeArgumentRefs maps the position of an element in the document (see
// elementPosition) to the eClassifier references of the type arguments of
// its eGenericType, "" for an argument without classifier (a wildcard or a
// type parameter).
type typeArgumentRefs map[string][]string

// elementPosition is the position of an element in an .ecore document: the
// local names of its ancestors and of itself, each with its index among the
// children of the same name of its parent ("/eClassifiers.3/eOperations.0").
func elementPosition(parent, local string, index int) string {
	return parent + "/" + local + "." + strconv.Itoa(index)
}

// stripGenericsWithArguments is stripGenerics, and returns the type arguments
// of the generic types it removed, which the languages with generics write
// (EList<EObject>).
func stripGenericsWithArguments(data []byte) ([]byte, typeArgumentRefs, error) {
	type frame struct {
		name     xml.Name
		insertAt int  // where an attribute is inserted in the start tag
		hasEType bool // the element has an eType attribute
		position string
		children map[string]int
	}
	type edit struct {
		start, end int // the bytes replaced
		text       string
	}
	var edits []edit
	var stack []*frame
	arguments := typeArgumentRefs{}
	skipDepth, skipStart := 0, 0
	// genericOwner is the position of the element whose eGenericType is
	// skipped, "" when the skipped tree is not an eGenericType.
	genericOwner, genericArguments := "", []string(nil)
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		start := int(decoder.InputOffset())
		token, err := decoder.RawToken()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		end := int(decoder.InputOffset())
		switch token := token.(type) {
		case xml.StartElement:
			if skipDepth > 0 {
				if skipDepth == 1 && genericOwner != "" && token.Name.Local == "eTypeArguments" {
					genericArguments = append(genericArguments, attributeValue(token, "eClassifier"))
				}
				skipDepth++
				continue
			}
			switch token.Name.Local {
			case "eGenericType":
				if len(stack) == 0 {
					return nil, nil, errors.New("malformed document: eGenericType at the root")
				}
				parent := stack[len(stack)-1]
				if !parent.hasEType {
					if classifier := attributeValue(token, "eClassifier"); classifier != "" {
						edits = append(edits, edit{start: parent.insertAt, end: parent.insertAt, text: ` eType="` + classifier + `"`})
						parent.hasEType = true
					}
				}
				skipDepth, skipStart = 1, start
				genericOwner, genericArguments = parent.position, nil
			case "eTypeParameters":
				skipDepth, skipStart = 1, start
				genericOwner = ""
			default:
				// the attribute goes before the '>' or the '/>' closing the start tag
				insertAt := end - 1
				if data[end-2] == '/' {
					insertAt = end - 2
				}
				position := ""
				if len(stack) > 0 {
					parent := stack[len(stack)-1]
					position = elementPosition(parent.position, token.Name.Local, parent.children[token.Name.Local])
					parent.children[token.Name.Local]++
				}
				stack = append(stack, &frame{name: token.Name, insertAt: insertAt, hasEType: attributeValue(token, "eType") != "",
					position: position, children: map[string]int{}})
			}
		case xml.EndElement:
			if skipDepth > 0 {
				if skipDepth--; skipDepth == 0 {
					edits = append(edits, edit{start: skipStart, end: end})
					if genericOwner != "" && len(genericArguments) > 0 {
						arguments[genericOwner] = genericArguments
					}
					genericOwner = ""
				}
				continue
			}
			// RawToken does not check the nesting
			if len(stack) == 0 || stack[len(stack)-1].name != token.Name {
				return nil, nil, errors.New("malformed document: unexpected end of element " + token.Name.Local)
			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) != 0 {
		return nil, nil, errors.New("malformed document: unclosed element " + stack[len(stack)-1].name.Local)
	}
	if len(edits) == 0 {
		return data, arguments, nil
	}
	sort.SliceStable(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	var result bytes.Buffer
	result.Grow(len(data))
	position := 0
	for _, e := range edits {
		result.Write(data[position:e.start])
		result.WriteString(e.text)
		position = e.end
	}
	result.Write(data[position:])
	return result.Bytes(), arguments, nil
}

// attributeValue returns the value of the attribute named name of element, ""
// when absent.
func attributeValue(element xml.StartElement, name string) string {
	for _, attr := range element.Attr {
		if attr.Name.Local == name && attr.Name.Space == "" {
			return attr.Value
		}
	}
	return ""
}
