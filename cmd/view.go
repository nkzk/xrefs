package view

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
	"github.com/nkzk/xrefs/internal/k8s"
	"github.com/nkzk/xrefs/internal/models"
	"github.com/nkzk/xrefs/internal/ui"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

type Cmd struct {
	Resource  string `required:"" name:"resource" arg:"" help:"The resource to view refs of, in the format 'TYPE[.VERSION][.GROUP][/NAME]'. required" xor:"resource,development"`
	Name      string `default:"" name:"name" arg:"" help:"resource name. optional" group:"resource"`
	Namespace string `default:"" name:"namespace" help:"resource namespace" group:"resource" short:"n"`

	KubeConfig string `default:"" help:"kubernetes kubeconfig location" name:"kube-config"`
	Context    string `default:"" help:"kubernetes context" name:"context" short:"c"`

	CacheOnDisk bool `help:"enable kubernetes discovery client caching to file instead of memory"`

	Mock bool `default:"false" help:"mock mode for development" group:"development" xor:"resource,development"`
}

func (c *Cmd) Help() string {
	return `
	This command will display and let you navigate the sub-resources of the targeted kubernetes resource
	
	Example usage:
	  xrefs view <kind>.<version>.<api-group>/<name>

	  xrefs view my-xr.alphav1.example.io/name
	`
}

func (c *Cmd) Run(k *kong.Context) error {
	ctx := context.Background()

	if c.Mock {
		return c.runMock(ctx, k)
	}

	return c.runKubernetes(ctx, k)
}

func (c *Cmd) runMock(ctx context.Context, k *kong.Context) error {
	kClient := k8s.NewMockClient()
	watcher := k8s.NewMockResourceWatcher()

	rootRef := &corev1.ObjectReference{
		APIVersion: "example.io/v1alpha1",
		Kind:       "MyXR",
		Name:       "example",
		Namespace:  "default",
	}

	if c.Resource == "kustomization" {
		rootRef = &corev1.ObjectReference{
			APIVersion: "kustomize.toolkit.fluxcd.io/v1",
			Kind:       "Kustomization",
			Name:       "example",
			Namespace:  "default",
		}
	}

	root, err := kClient.GetUnstructured(ctx, rootRef)
	if err != nil {
		return err
	}

	rootResource := models.NewResource(
		nil,
		root,
		rootRef,
	)
	rootResource.Expanded = true

	return c.watchResourceTree(ctx, k, kClient, watcher, rootResource)
}

func (c *Cmd) runKubernetes(ctx context.Context, k *kong.Context) error {
	clientconfig, client, rmapper, err := k8s.SetupKubeClient(c.KubeConfig, c.Context, c.CacheOnDisk)
	if err != nil {
		return err
	}

	resource, name, err := k8s.ParseResourceName(c.Resource, c.Name)
	if err != nil {
		return err
	}

	resourceMapping, err := k8s.MappingFor(rmapper, resource)
	if err != nil {
		return err
	}

	resourceObjectRef, err := k8s.ResourceObjectRefFromMapping(
		resourceMapping,
		clientconfig,
		name,
		c.Namespace,
	)
	if err != nil {
		return err
	}

	kClient := k8s.NewK8sClient(client)
	watcher := k8s.NewKubernetesResourceWatcher(client)

	root, err := kClient.GetUnstructured(ctx, resourceObjectRef)
	if err != nil {
		return err
	}

	rootResource := models.NewResource(
		nil,
		root,
		resourceObjectRef,
	)
	rootResource.Expanded = true

	return c.watchResourceTree(ctx, k, kClient, watcher, rootResource)
}

func (c *Cmd) watchResourceTree(
	ctx context.Context,
	k *kong.Context,
	kClient k8s.Client,
	watcher k8s.ResourceWatcher,
	root *models.Resource,
) error {
	w, err := watcher.WatchResource(ctx, root)
	if err != nil {
		return fmt.Errorf("cannot start watch: %v", err)
	}
	defer w.Stop()

	prog := tea.NewProgram(ui.NewModel(root), tea.WithOutput(k.Stdout))

	go c.watchProducer(ctx, kClient, root, prog, w)

	_, err = prog.Run()
	return err
}

// runs the watchProducer loop for a resource and sends updates to bubbletea tui
func (c *Cmd) watchProducer(ctx context.Context, kClient k8s.Client, root *models.Resource, prog *tea.Program, w watch.Interface) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	if err := update(ctx, root, kClient, prog); err != nil {
		c.handleProducerError(prog, err)

		return
	}
	prog.Send(ui.UpdateResourceMsg{
		Resource: root,
	})

	// watch loop
	for {
		select {
		case evt, ok := <-w.ResultChan():
			if !ok {
				prog.Send(ui.QuitMsg{})
				return
			}
			if evt.Type == watch.Deleted {
				prog.Send(ui.RootDeletedMsg{})
				return
			}
			if err := update(ctx, root, kClient, prog); err != nil {
				c.handleProducerError(prog, err)
				return
			}
			prog.Send(ui.UpdateResourceMsg{
				Resource: root,
			})

		case <-ticker.C:
			if err := update(ctx, root, kClient, prog); err != nil {
				c.handleProducerError(prog, err)
				return
			}

			prog.Send(ui.UpdateResourceMsg{
				Resource: root,
			})
		case <-ctx.Done():
			prog.Send(ui.QuitMsg{})
			return
		}
	}
}

// updateAndSend updates a Resource and its children and send an UpdateResourceMsg to tea.Program
func update(ctx context.Context, r *models.Resource, kClient k8s.Client, prog *tea.Program) error {
	current, err := kClient.GetUnstructured(ctx, r.Ref)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.NotFound = true
			return nil
		}

		return err
	}

	r.NotFound = false

	// Update unstructured
	r.Unstructured = current

	// Update Resource conditions
	freshConditions := []models.Condition{}
	conditions, ok, err := unstructured.NestedSlice(current.Object, "status", "conditions")
	if ok && err == nil {
		for _, c := range conditions {
			var condition models.Condition
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
				c.(map[string]any),
				&condition,
			); err == nil {
				freshConditions = append(freshConditions, condition)
			}
		}
	}

	r.Conditions = freshConditions

	loadResourceChildren(r)

	if !r.Expanded {
		return nil
	}

	r.ChildrenLoaded = true

	// Update children
	for i := range r.Children {
		if err := update(ctx, &r.Children[i], kClient, prog); err != nil {
			return err
		}
	}

	return nil
}

// loads resource-refs of a root resource to the Children array.
func loadResourceChildren(root *models.Resource) {
	existingChildren := make(map[string]*models.Resource)
	for i := range root.Children {
		c := &root.Children[i]
		if c.Ref != nil {
			key := fmt.Sprintf("%s/%s/%s", c.Ref.APIVersion, c.Ref.Kind, c.Ref.Name)
			existingChildren[key] = c
		}
	}

	var newChildren []models.Resource

	switch root.Unstructured.GroupVersionKind() {
	case schema.GroupVersionKind{
		Group:   "kustomize.toolkit.fluxcd.io",
		Version: "v1",
		Kind:    "Kustomization"}:
		entries, ok, err := unstructured.NestedSlice(root.Unstructured.Object, "status", "inventory", "entries")
		if err != nil || !ok {
			root.Children = nil
			return
		}

		for _, e := range entries {
			m, ok := e.(map[string]any)
			if !ok {
				continue
			}

			// id format: <namespace>_<name>_<group>_<kind>
			// "v" for version
			id, _ := m["id"].(string)
			version, _ := m["v"].(string)

			parts := strings.SplitN(id, "_", 4)
			if len(parts) < 4 {
				continue
			}

			ns, name, group, kind := parts[0], parts[1], parts[2], parts[3]

			apiVersion := version
			if group != "" {
				apiVersion = group + "/" + version
			}

			ref := &corev1.ObjectReference{
				APIVersion: apiVersion,
				Kind:       kind,
				Name:       name,
				Namespace:  ns,
			}

			newChildren = append(newChildren, *models.NewResource(nil, nil, ref))
		}
	default:
		// assume crossplane XR

		resourceRefs, ok, err := unstructured.NestedSlice(root.Unstructured.Object, "spec", "crossplane", "resourceRefs")
		if err != nil || !ok {
			root.Children = nil
			return
		}

		parentNS := root.Unstructured.GetNamespace()

		for _, r := range resourceRefs {
			ref := corev1.ObjectReference{}

			m, ok := r.(map[string]any)
			if !ok {
				continue
			}

			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(m, &ref); err != nil {
				continue
			}

			if ref.Namespace == "" && parentNS != "" {
				ref.Namespace = parentNS
			}

			newChildren = append(newChildren, *models.NewResource(nil, nil, &ref))
		}
	}

	// Merge: preserve Expanded/ChildrenLoaded/Children state from existing children
	for i := range newChildren {
		ref := newChildren[i].Ref
		if ref != nil {
			key := fmt.Sprintf("%s/%s/%s", ref.APIVersion, ref.Kind, ref.Name)
			if existing, ok := existingChildren[key]; ok {
				newChildren[i].Expanded = existing.Expanded
				newChildren[i].ChildrenLoaded = existing.ChildrenLoaded
				newChildren[i].Children = existing.Children
				newChildren[i].Unstructured = existing.Unstructured
				newChildren[i].Conditions = existing.Conditions
				newChildren[i].Error = existing.Error
				newChildren[i].NotFound = existing.NotFound
			}
		}
	}

	root.Children = newChildren
}

// handleProducerError handles errors from the watch producer.
func (c *Cmd) handleProducerError(prog *tea.Program, err error) {
	if apierrors.IsNotFound(err) {
		return
	}

	prog.Send(ui.RootErrMsg{Err: fmt.Errorf("error getting resource: %v", err)})
}
