package ui

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/nkzk/xrefs/internal/models"
)

func flatten(r models.Resource, depth int) []list.Item {
	return flattenWithPrefix(r, depth, true, "")
}

func flattenWithPrefix(r models.Resource, depth int, isLast bool, prefix string) []list.Item {
	r.Depth = depth
	r.IsLast = isLast
	r.Prefix = prefix

	out := []list.Item{r}

	childPrefix := prefix
	if depth > 0 {
		if isLast {
			childPrefix += "   "
		} else {
			childPrefix += "│  "
		}
	}

	for i, child := range r.Children {
		out = append(out, flattenWithPrefix(
			child,
			depth+1,
			i == len(r.Children)-1,
			childPrefix,
		)...)
	}

	return out
}

func treeName(r models.Resource) string {
	prefix := ""
	if r.Depth > 0 {
		if r.IsLast {
			prefix = r.Prefix + "└─ "
		} else {
			prefix = r.Prefix + "├─ "
		}
	}

	kind := r.Ref.Kind
	name := r.Ref.Name

	if r.Unstructured != nil {
		if r.Unstructured.GetKind() != "" {
			kind = r.Unstructured.GetKind()
		}
		if r.Unstructured.GetName() != "" {
			name = r.Unstructured.GetName()
		}
	}

	return fmt.Sprintf("%s%s/%s", prefix, kind, name)
}

func namespace(r models.Resource) string {
	if r.Unstructured != nil && r.Unstructured.GetNamespace() != "" {
		return r.Unstructured.GetNamespace()
	}
	if r.Ref != nil {
		return r.Ref.Namespace
	}
	return "-"
}

func condStatus(r models.Resource, name string) string {
	c := r.Conditions.Get(name)
	if c.Status == "" {
		return "-"
	}
	return c.Status
}

func condReason(r models.Resource) string {
	if r := r.Conditions.Get("Ready").Reason; r != "" {
		return r
	}
	if r := r.Conditions.Get("Synced").Reason; r != "" {
		return r
	}
	return "-"
}

func shorten(s string, max int) string {
	if len(s) <= max {
		return s
	}
	head := max / 2
	tail := max - head - 1
	return s[:head] + "(...)" + s[len(s)-tail:]
}
