package ui

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/nkzk/xrefs/internal/models"
)

type ResourceItem struct {
	Resource *models.Resource
	Prefix   string
}

func (i ResourceItem) Title() string       { return resourceName(i.Resource) }
func (i ResourceItem) Description() string { return resourceNamespace(i.Resource) }
func (i ResourceItem) FilterValue() string { return resourceName(i.Resource) }

func visibleResourceItems(root *models.Resource) []list.Item {
	var items []list.Item

	var walk func(r *models.Resource, rowPrefix, indent string)
	walk = func(r *models.Resource, rowPrefix, indent string) {
		if r == nil {
			return
		}

		items = append(items, ResourceItem{
			Resource: r,
			Prefix:   rowPrefix,
		})

		if !r.Expanded {
			return
		}

		for i := range r.Children {
			last := i == len(r.Children)-1

			connector := "├─ "
			nextIndent := indent + "│  "
			if last {
				connector = "└─ "
				nextIndent = indent + "   "
			}

			walk(&r.Children[i], indent+connector, nextIndent)
		}
	}

	walk(root, "", "")
	return items
}

func renderResourceName(item ResourceItem) string {
	r := item.Resource
	label := resourceName(r)

	if r != nil && len(r.Children) > 0 {
		if r.Expanded {
			label = "▼ " + label
		} else {
			label = "▶ " + label
		}
	}

	const maxCol = 64
	maxLabel := maxCol - len(item.Prefix)
	if maxLabel <= 0 {
		return shorten(item.Prefix, maxCol)
	}

	return item.Prefix + shorten(label, maxLabel)
}

func resourceName(r *models.Resource) string {
	if r == nil {
		return "-"
	}

	kind, name := "", ""
	if r.Ref != nil {
		kind, name = r.Ref.Kind, r.Ref.Name
	}
	if r.Unstructured != nil {
		if k := r.Unstructured.GetKind(); k != "" {
			kind = k
		}
		if n := r.Unstructured.GetName(); n != "" {
			name = n
		}
	}

	if kind == "" {
		kind = "?"
	}
	if name == "" {
		name = "?"
	}
	return fmt.Sprintf("%s/%s", kind, name)
}

func resourceNamespace(r *models.Resource) string {
	if r == nil {
		return "-"
	}
	if r.Unstructured != nil && r.Unstructured.GetNamespace() != "" {
		return r.Unstructured.GetNamespace()
	}
	if r.Ref != nil && r.Ref.Namespace != "" {
		return r.Ref.Namespace
	}
	return "-"
}

func condStatus(r *models.Resource, name string) string {
	if r == nil {
		return "-"
	}

	c := r.Conditions.Get(name)
	if c.Status == "" {
		return "-"
	}
	return c.Status
}

func condReason(r *models.Resource) string {
	if r == nil {
		return "-"
	}
	if reason := r.Conditions.Get("Ready").Reason; reason != "" {
		return reason
	}
	if reason := r.Conditions.Get("Synced").Reason; reason != "" {
		return reason
	}
	return "-"
}
