package store

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"

	"github.com/nkzk/xrefs/internal/ui/v2/messages"
	"github.com/nkzk/xrefs/internal/utils"
)

type Config struct {
	ResourceName    string
	ResourceGroup   string
	ResourceVersion string
	Name            string
	Namespace       string
}

type Store struct {
	config Config
	client Client

	statusCache sync.Map

	updating bool
}

type Client interface {
	Run(command string) (string, error)
}

type KubectlClient struct{}

func (k KubectlClient) Run(command string) (string, error) {
	args := strings.Fields(command)
	if len(args) == 0 {
		return "", fmt.Errorf("empty command")
	}
	output, err := utils.RunCommand(args[0], args[1:]...)
	return string(output), err
}

func NewStore(cfg Config, client Client) *Store {
	return &Store{
		config: cfg,
		client: client,
	}
}

func NewKubectlClient() *KubectlClient {
	return &KubectlClient{}
}

func (s *Store) FetchResources() tea.Cmd {
	return func() tea.Msg {
		cmd, err := s.buildGetCommand()
		if err != nil {
			return messages.ErrMsg{Err: err}
		}

		output, err := s.client.Run(cmd)
		if err != nil {
			return messages.ErrMsg{Err: fmt.Errorf("kubectl failed: %w", err)}
		}

		resources, err := s.parseXR(output)
		if err != nil {
			return messages.ErrMsg{Err: err}
		}

		return messages.ResourcesLoadedMsg{Resources: resources}
	}
}

func (s *Store) UpdateStatus(resources []messages.Resource) tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup

		for _, r := range resources {
			wg.Add(1)
			go func(res messages.Resource) {
				defer wg.Done()
				s.fetchResourceStatus(res)
			}(r)
		}

		wg.Wait()

		updated := make([]messages.Resource, len(resources))
		for i, r := range resources {
			updated[i] = s.enrichWithStatus(r)
		}

		return messages.ResourceStatusUpdatedMsg{Resources: updated}
	}
}

func (s *Store) FetchYAML(r messages.Resource) tea.Cmd {
	return func() tea.Msg {
		cmd, err := s.buildResourceCommand(r, "yaml")
		if err != nil {
			return messages.ErrMsg{Err: err}
		}

		output, err := s.client.Run(cmd)
		if err != nil {
			return messages.ErrMsg{Err: err}
		}

		return messages.ContentLoadedMsg{Content: output}
	}
}

func (s *Store) FetchDescribe(r messages.Resource) tea.Cmd {
	return func() tea.Msg {
		cmd, err := s.buildDescribeCommand(r)
		if err != nil {
			return messages.ErrMsg{Err: err}
		}

		output, err := s.client.Run(cmd)
		if err != nil {
			return messages.ErrMsg{Err: err}
		}

		return messages.ContentLoadedMsg{Content: output}
	}
}

func Tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return messages.TickMsg(t)
	})
}

func (s *Store) buildGetCommand() (string, error) {
	cfg := s.config
	if cfg.ResourceName == "" || cfg.ResourceVersion == "" || cfg.Name == "" || cfg.Namespace == "" {
		return "", fmt.Errorf("missing required config fields")
	}

	apiVersion := cfg.ResourceVersion
	if i := strings.IndexRune(apiVersion, '/'); i != -1 {
		apiVersion = apiVersion[:i]
		return fmt.Sprintf("kubectl get %s.%s/%s -n %s -o yaml",
			cfg.ResourceName, apiVersion, cfg.Name, cfg.Namespace), nil
	}

	return fmt.Sprintf("kubectl get %s.%s.%s/%s -n %s -o yaml",
		cfg.ResourceName, apiVersion, cfg.ResourceGroup, cfg.Name, cfg.Namespace), nil
}

func (s *Store) buildResourceCommand(r messages.Resource, format string) (string, error) {
	if r.Kind == "" || r.APIVersion == "" || r.Name == "" || r.Namespace == "" {
		return "", fmt.Errorf("missing required resource fields")
	}

	apiVersion := r.APIVersion
	if i := strings.IndexRune(apiVersion, '/'); i != -1 {
		apiVersion = apiVersion[:i]
		return fmt.Sprintf("kubectl get %s.%s/%s -n %s -o %s",
			r.Kind, apiVersion, r.Name, r.Namespace, format), nil
	}

	return fmt.Sprintf("kubectl get %s.%s/%s -n %s -o %s",
		r.Kind, apiVersion, r.Name, r.Namespace, format), nil
}

func (s *Store) buildDescribeCommand(r messages.Resource) (string, error) {
	if r.Kind == "" || r.APIVersion == "" || r.Name == "" || r.Namespace == "" {
		return "", fmt.Errorf("missing required resource fields")
	}

	apiVersion := r.APIVersion
	if i := strings.IndexRune(apiVersion, '/'); i != -1 {
		apiVersion = apiVersion[:i]
		return fmt.Sprintf("kubectl describe %s.%s/%s -n %s",
			r.Kind, apiVersion, r.Name, r.Namespace), nil
	}

	return fmt.Sprintf("kubectl describe %s.%s/%s -n %s",
		r.Kind, apiVersion, r.Name, r.Namespace), nil
}

func (s *Store) parseXR(yamlStr string) ([]messages.Resource, error) {
	var xr struct {
		Metadata struct {
			Namespace string `yaml:"namespace"`
		} `yaml:"metadata"`
		Spec struct {
			Crossplane struct {
				ResourceRefs []struct {
					APIVersion string `yaml:"apiVersion"`
					Kind       string `yaml:"kind"`
					Name       string `yaml:"name"`
				} `yaml:"resourceRefs"`
			} `yaml:"crossplane"`
		} `yaml:"spec"`
	}

	if err := yaml.Unmarshal([]byte(yamlStr), &xr); err != nil {
		return nil, fmt.Errorf("failed to parse XR: %w", err)
	}

	resources := make([]messages.Resource, 0, len(xr.Spec.Crossplane.ResourceRefs))
	for _, ref := range xr.Spec.Crossplane.ResourceRefs {
		resources = append(resources, messages.Resource{
			Namespace:    xr.Metadata.Namespace,
			Kind:         ref.Kind,
			APIVersion:   ref.APIVersion,
			Name:         ref.Name,
			Synced:       "-",
			SyncedReason: "-",
			Ready:        "-",
			ReadyReason:  "-",
		})
	}

	return resources, nil
}

func (s *Store) fetchResourceStatus(r messages.Resource) {
	cmd, err := s.buildResourceCommand(r, "yaml")
	if err != nil {
		return
	}

	output, err := s.client.Run(cmd)
	if err != nil {
		return
	}

	status, err := s.parseStatus(output)
	if err != nil {
		return
	}

	s.statusCache.Store(r.Name, status)
}

func (s *Store) enrichWithStatus(r messages.Resource) messages.Resource {
	if cached, ok := s.statusCache.Load(r.Name); ok {
		if status, ok := cached.(resourceStatus); ok {
			r.Synced = status.getCondition("Synced").Status
			r.SyncedReason = status.getCondition("Synced").Reason
			r.Ready = status.getCondition("Ready").Status
			r.ReadyReason = status.getCondition("Ready").Reason
		}
	}
	return r
}

type resourceStatus struct {
	Conditions []condition `yaml:"conditions"`
}

type condition struct {
	Type   string `yaml:"type"`
	Status string `yaml:"status"`
	Reason string `yaml:"reason"`
}

func (s resourceStatus) getCondition(t string) condition {
	for _, c := range s.Conditions {
		if c.Type == t {
			return c
		}
	}
	return condition{Status: "-", Reason: "-"}
}

func (s *Store) parseStatus(yamlStr string) (resourceStatus, error) {
	var parsed struct {
		Status resourceStatus `yaml:"status"`
	}

	if err := yaml.Unmarshal([]byte(yamlStr), &parsed); err != nil {
		return resourceStatus{}, err
	}

	return parsed.Status, nil
}
