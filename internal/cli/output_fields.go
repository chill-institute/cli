package cli

import (
	"regexp"
	"strings"
)

var fieldSegmentPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type fieldSelection struct {
	root *fieldNode
}

type fieldNode struct {
	include  bool
	children map[string]*fieldNode
}

func parseFieldSelection(raw string) (*fieldSelection, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	root := &fieldNode{children: map[string]*fieldNode{}}
	for _, path := range strings.Split(trimmed, ",") {
		normalizedPath := strings.TrimSpace(path)
		if normalizedPath == "" {
			return nil, usageError("invalid_fields", "--fields contains an empty path")
		}

		node := root
		for _, segment := range strings.Split(normalizedPath, ".") {
			normalizedSegment := strings.TrimSpace(segment)
			if normalizedSegment == "" {
				return nil, usageError("invalid_fields", "invalid field path %q", normalizedPath)
			}
			if !fieldSegmentPattern.MatchString(normalizedSegment) {
				return nil, usageError("invalid_fields", "invalid field segment %q in %q", normalizedSegment, normalizedPath)
			}
			if node.children == nil {
				node.children = map[string]*fieldNode{}
			}
			child := node.children[normalizedSegment]
			if child == nil {
				child = &fieldNode{}
				node.children[normalizedSegment] = child
			}
			node = child
		}
		node.include = true
	}

	return &fieldSelection{root: root}, nil
}

func (selection *fieldSelection) apply(value any) any {
	if selection == nil || selection.root == nil {
		return value
	}
	filtered, ok := selection.root.apply(value)
	if ok {
		return filtered
	}
	return map[string]any{}
}

func (node *fieldNode) apply(value any) (any, bool) {
	if node == nil {
		return value, true
	}
	if node.include {
		return value, true
	}

	switch typed := value.(type) {
	case map[string]any:
		filtered := make(map[string]any, len(node.children))
		for key, child := range node.children {
			nextValue, ok := typed[key]
			if !ok {
				continue
			}
			applied, appliedOK := child.apply(nextValue)
			if appliedOK {
				filtered[key] = applied
			}
		}
		return filtered, true
	case []any:
		filtered := make([]any, 0, len(typed))
		for _, item := range typed {
			applied, appliedOK := node.apply(item)
			if appliedOK {
				filtered = append(filtered, applied)
			}
		}
		return filtered, true
	default:
		return nil, false
	}
}
