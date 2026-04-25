package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
	"github.com/google/uuid"
	"github.com/nkzk/xrefs/internal/k8s"
	"github.com/nkzk/xrefs/internal/models"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type Cmd struct {
	Resource  string `required:"" name:"resource" arg:"" help:"The resource to view refs of, in the format 'TYPE[.VERSION][.GROUP][/NAME]'. required" xor:"resource,development"`
	Name      string `default:"" name:"name" arg:"" help:"resource name. optional" group:"resource"`
	Namespace string `default:"" name:"namespace" help:"resource namespace" group:"resource" short:"n"`

	Context     string `default:"" help:"kubernetes context" name:"context" short:"c"`
	CacheOnDisk bool   `help:"enable kubernetes discovery client caching to file instead of memory"`

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
	clientconfig, client, rmapper, err := k8s.SetupKubeClient(c.Context, c.CacheOnDisk)
	if err != nil {
		return err
	}

	// since resource name can come as its own arg or part of resource-arg, parse it
	resource, name, err := k8s.ParseResourceName(c.Resource, c.Name)
	if err != nil {
		return err
	}

	resourceMapping, err := k8s.MappingFor(rmapper, resource)
	if err != nil {
		return err
	}

	resourceoObjectRef, err := k8s.ResourceObjectFromMapping(resourceMapping, clientconfig, name, c.Namespace)
	if err != nil {
		return err
	}

	kClient := k8s.NewK8sClient(client)

	root := kClient.GetResource(context.TODO(), resourceoObjectRef)

	c.initChildren(context.TODO(), root, resourceMapping, kClient)

	m := models.NewModel(root)

	_, err = tea.NewProgram(m).Run()
	if err != nil {
		fmt.Fprintf(k.Stdout, "failed to start: %v\n", err)
		os.Exit(1)
	}

	return nil
}

// runs the watch loop and sends updates to bubbletea tui
func (c *Cmd) watch(ctx context.Context, kClient k8s.Client, root *v1.ObjectReference, mapping *meta.RESTMapping, prog *tea.Program, w watch.Interface) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	last := &models.Resource{}

	renderAndSend := func() error {
		current := kClient.GetResource(ctx, root)
		if current.Error != nil {
			if apierrors.IsNotFound(current.Error) {
				prog.Send(models.RootNotFoundMsg{})
				return nil
			}

			return current.Error
		}

		if !reflect.DeepEqual(current, last) {
			prog.Send(models.UpdateResourceMsg{
				Resource: last,
			})
		}

		return nil
	}
	// watch loop
	for {
		select {
		case evt, ok := <-w.ResultChan():
			if !ok {
				prog.Send(models.QuitMsg{})
				return
			}
			if evt.Type == watch.Deleted {
				prog.Send(models.RootDeletedMsg{})
				return
			}
			if err := renderAndSend(); err != nil {
				c.handleProducerError(prog, err)
				return
			}

		case <-ticker.C:
			if err := renderAndSend(); err != nil {
				c.handleProducerError(prog, err)
				return
			}
		case <-ctx.Done():
			prog.Send(models.QuitMsg{})
			return
		}
	}
}

func (c *Cmd) initChildren(ctx context.Context, root *models.Resource, mapping *meta.RESTMapping, kClient k8s.Client) {
	// reset the array to ensure we start fresh
	root.Children = []models.Resource{}

	switch root.Unstructured.GroupVersionKind() {
	// assume resource is a crossplane XR
	default:
		resourceRefs, ok, err := unstructured.NestedSlice(root.Unstructured.Object, "spec", "crossplane", "resourceRefs")
		if err != nil || !ok {
			return
		}

		for _, r := range resourceRefs {
			ref, ok := r.(corev1.ObjectReference)
			if !ok {
				continue
			}

			u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&ref)
			if err != nil {
				continue
			}

			root.Children = append(root.Children, models.Resource{
				Parent: root,
				ID:     uuid.New().String(),
				Unstructured: unstructured.Unstructured{
					Object: u,
				},
			})
		}
		// other resources can be represented as a case, for example flux kustomization
	}
}

// handleProducerError handles errors from the watch producer.
func (c *Cmd) handleProducerError(prog *tea.Program, err error) {
	if apierrors.IsNotFound(err) {
		prog.Send(models.RootNotFoundMsg{})
		return
	}
	prog.Send(models.RootErrMsg{Err: fmt.Errorf("error getting resource: %v", err)})
}
