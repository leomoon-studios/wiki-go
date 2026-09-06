package handlers

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

const (
	svgNamespace      = "http://www.w3.org/2000/svg"
	xlinkNamespace    = "http://www.w3.org/1999/xlink"
	maximumSVGDepth   = 64
	maximumSVGAttrs   = 128
	maximumSVGTextLen = 1 << 20
)

// The element allowlist is limited to static vector geometry, text, and
// same-document paint definitions. It deliberately excludes scripting,
// animation, stylesheets, links, images, use, and foreign/embedded content.
var allowedSVGElements = map[string]struct{}{
	"svg": {}, "g": {}, "defs": {},
	"path": {}, "rect": {}, "circle": {}, "ellipse": {}, "line": {}, "polyline": {}, "polygon": {},
	"text": {}, "tspan": {}, "title": {}, "desc": {},
	"linearGradient": {}, "radialGradient": {}, "stop": {},
	"clipPath": {}, "mask": {}, "marker": {}, "pattern": {},
}

// The attribute allowlist contains geometry and presentation values only. URL
// attributes, event handlers, style/class hooks, and arbitrary namespaces are
// excluded; reference-bearing presentation attributes are validated below.
var allowedSVGAttributes = map[string]struct{}{
	"id": {}, "version": {}, "role": {}, "aria-label": {}, "aria-hidden": {},
	"width": {}, "height": {}, "viewBox": {}, "preserveAspectRatio": {},
	"x": {}, "y": {}, "x1": {}, "y1": {}, "x2": {}, "y2": {},
	"cx": {}, "cy": {}, "r": {}, "rx": {}, "ry": {}, "d": {}, "points": {}, "pathLength": {},
	"dx": {}, "dy": {}, "rotate": {}, "textLength": {}, "lengthAdjust": {},
	"transform": {}, "opacity": {}, "visibility": {}, "display": {},
	"fill": {}, "fill-opacity": {}, "fill-rule": {},
	"stroke": {}, "stroke-opacity": {}, "stroke-width": {}, "stroke-linecap": {}, "stroke-linejoin": {},
	"stroke-miterlimit": {}, "stroke-dasharray": {}, "stroke-dashoffset": {},
	"clip-path": {}, "clip-rule": {}, "mask": {},
	"color": {}, "vector-effect": {}, "shape-rendering": {},
	"font-family": {}, "font-size": {}, "font-style": {}, "font-weight": {}, "letter-spacing": {},
	"text-anchor": {}, "dominant-baseline": {},
	"offset": {}, "stop-color": {}, "stop-opacity": {},
	"gradientUnits": {}, "gradientTransform": {}, "spreadMethod": {}, "fx": {}, "fy": {}, "fr": {},
	"clipPathUnits": {}, "maskUnits": {}, "maskContentUnits": {},
	"markerWidth": {}, "markerHeight": {}, "refX": {}, "refY": {}, "orient": {}, "markerUnits": {},
	"marker-start": {}, "marker-mid": {}, "marker-end": {},
	"patternUnits": {}, "patternContentUnits": {}, "patternTransform": {},
}

var svgInternalReferenceAttributes = map[string]struct{}{
	"clip-path": {}, "mask": {}, "marker-start": {}, "marker-mid": {}, "marker-end": {},
}

var svgPaintAttributes = map[string]struct{}{
	"fill": {}, "stroke": {}, "stop-color": {}, "color": {},
}

var svgTextElements = map[string]struct{}{
	"text": {}, "tspan": {}, "title": {}, "desc": {},
}

// sanitizeSVG parses and re-encodes a deliberately small, static subset of
// SVG. Unknown markup is rejected instead of being copied or regex-rewritten.
func sanitizeSVG(content []byte) ([]byte, error) {
	decoder := xml.NewDecoder(bytes.NewReader(content))
	decoder.Strict = true

	var output bytes.Buffer
	encoder := xml.NewEncoder(&output)
	stack := make([]string, 0, 8)
	rootSeen := false
	rootClosed := false
	textLength := 0

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid SVG XML: %w", err)
		}

		switch typed := token.(type) {
		case xml.StartElement:
			if rootClosed {
				return nil, fmt.Errorf("multiple SVG root elements")
			}
			if len(stack) >= maximumSVGDepth {
				return nil, fmt.Errorf("SVG nesting exceeds %d elements", maximumSVGDepth)
			}
			if typed.Name.Space != "" && typed.Name.Space != svgNamespace {
				return nil, fmt.Errorf("disallowed SVG element namespace %q", typed.Name.Space)
			}
			if !rootSeen {
				if typed.Name.Local != "svg" {
					return nil, fmt.Errorf("SVG root element is required")
				}
				rootSeen = true
			}
			if _, allowed := allowedSVGElements[typed.Name.Local]; !allowed {
				return nil, fmt.Errorf("disallowed SVG element %q", typed.Name.Local)
			}

			attributes, err := sanitizeSVGAttributes(typed, len(stack) == 0)
			if err != nil {
				return nil, err
			}
			clean := xml.StartElement{Name: xml.Name{Local: typed.Name.Local}, Attr: attributes}
			if err := encoder.EncodeToken(clean); err != nil {
				return nil, fmt.Errorf("encode SVG element: %w", err)
			}
			stack = append(stack, typed.Name.Local)

		case xml.EndElement:
			if len(stack) == 0 || stack[len(stack)-1] != typed.Name.Local {
				return nil, fmt.Errorf("invalid SVG element closure %q", typed.Name.Local)
			}
			name := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if err := encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: name}}); err != nil {
				return nil, fmt.Errorf("encode SVG end element: %w", err)
			}
			if len(stack) == 0 {
				rootClosed = true
			}

		case xml.CharData:
			text := string(typed)
			if len(stack) == 0 {
				if strings.TrimSpace(text) != "" {
					return nil, fmt.Errorf("text outside SVG root element")
				}
				continue
			}
			if strings.TrimSpace(text) != "" {
				if _, allowed := svgTextElements[stack[len(stack)-1]]; !allowed {
					return nil, fmt.Errorf("text is not allowed inside SVG element %q", stack[len(stack)-1])
				}
				textLength += len(text)
				if textLength > maximumSVGTextLen {
					return nil, fmt.Errorf("SVG text exceeds size limit")
				}
			}
			if err := encoder.EncodeToken(xml.CharData([]byte(text))); err != nil {
				return nil, fmt.Errorf("encode SVG text: %w", err)
			}

		case xml.Comment:
			// Comments carry no rendering information and are omitted.

		case xml.ProcInst:
			if typed.Target != "xml" || rootSeen || len(stack) != 0 {
				return nil, fmt.Errorf("processing instructions are not allowed in SVG")
			}
			// A leading XML declaration is accepted but omitted from canonical output.

		case xml.Directive:
			return nil, fmt.Errorf("directives and entities are not allowed in SVG")

		default:
			return nil, fmt.Errorf("unsupported SVG token")
		}
	}

	if !rootSeen || !rootClosed || len(stack) != 0 {
		return nil, fmt.Errorf("complete SVG root element is required")
	}
	if err := encoder.Flush(); err != nil {
		return nil, fmt.Errorf("flush SVG output: %w", err)
	}
	return output.Bytes(), nil
}

func sanitizeSVGAttributes(element xml.StartElement, root bool) ([]xml.Attr, error) {
	if len(element.Attr) > maximumSVGAttrs {
		return nil, fmt.Errorf("too many attributes on SVG element %q", element.Name.Local)
	}

	attributes := make([]xml.Attr, 0, len(element.Attr)+1)
	if root {
		attributes = append(attributes, xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: svgNamespace})
	}
	seen := make(map[string]struct{}, len(element.Attr))
	for _, attribute := range element.Attr {
		if isNamespaceDeclaration(attribute) {
			if attribute.Value != svgNamespace && attribute.Value != xlinkNamespace {
				return nil, fmt.Errorf("disallowed SVG namespace declaration %q", attribute.Value)
			}
			continue
		}
		if attribute.Name.Space != "" {
			return nil, fmt.Errorf("namespaced SVG attribute %q is not allowed", attribute.Name.Local)
		}

		name := attribute.Name.Local
		lowerName := strings.ToLower(name)
		if strings.HasPrefix(lowerName, "on") || lowerName == "style" || lowerName == "href" || lowerName == "src" {
			return nil, fmt.Errorf("active SVG attribute %q is not allowed", name)
		}
		if _, allowed := allowedSVGAttributes[name]; !allowed {
			return nil, fmt.Errorf("disallowed SVG attribute %q", name)
		}
		if _, duplicate := seen[name]; duplicate {
			return nil, fmt.Errorf("duplicate SVG attribute %q", name)
		}
		seen[name] = struct{}{}

		value := attribute.Value
		if _, internalOnly := svgInternalReferenceAttributes[name]; internalOnly {
			if value != "none" && !isInternalSVGReference(value) {
				return nil, fmt.Errorf("external SVG reference in attribute %q", name)
			}
		}
		if _, paint := svgPaintAttributes[name]; paint && !isSafeSVGPaint(value) {
			return nil, fmt.Errorf("invalid or external SVG paint in attribute %q", name)
		}

		attributes = append(attributes, xml.Attr{Name: xml.Name{Local: name}, Value: value})
	}
	return attributes, nil
}

func isSafeSVGPaint(value string) bool {
	value = strings.TrimSpace(value)
	if isInternalSVGReference(value) {
		return true
	}
	if value == "" {
		return false
	}
	if strings.HasPrefix(value, "#") {
		digits := value[1:]
		if len(digits) != 3 && len(digits) != 4 && len(digits) != 6 && len(digits) != 8 {
			return false
		}
		for _, character := range digits {
			if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
				return false
			}
		}
		return true
	}
	if strings.Contains(value, "(") {
		lower := strings.ToLower(value)
		allowedFunction := strings.HasPrefix(lower, "rgb(") || strings.HasPrefix(lower, "rgba(") ||
			strings.HasPrefix(lower, "hsl(") || strings.HasPrefix(lower, "hsla(")
		if !allowedFunction || !strings.HasSuffix(value, ")") {
			return false
		}
		for _, character := range value {
			if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
				(character >= '0' && character <= '9') || strings.ContainsRune("(),.% +-", character) {
				continue
			}
			return false
		}
		return true
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || character == '-' {
			continue
		}
		return false
	}
	return true
}

func isNamespaceDeclaration(attribute xml.Attr) bool {
	return (attribute.Name.Space == "" && attribute.Name.Local == "xmlns") || attribute.Name.Space == "xmlns"
}

func isInternalSVGReference(value string) bool {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "url(#") || !strings.HasSuffix(value, ")") {
		return false
	}
	identifier := strings.TrimSuffix(strings.TrimPrefix(value, "url(#"), ")")
	if identifier == "" {
		return false
	}
	for _, character := range identifier {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || strings.ContainsRune("_-:.", character) {
			continue
		}
		return false
	}
	return true
}
