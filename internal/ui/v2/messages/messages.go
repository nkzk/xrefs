package messages

import "time"

type TickMsg time.Time

type ErrMsg struct{ Err error }

func (e ErrMsg) Error() string { return e.Err.Error() }

type (
	ResourcesLoadedMsg struct {
		Resources []Resource
	}

	ResourceStatusUpdatedMsg struct {
		Resources []Resource
	}

	ContentLoadedMsg struct {
		Content string
	}
)

type ShowContentMsg struct {
	Resource Resource
	Mode     string
}

type GoBackMsg struct{}

type Resource struct {
	Namespace    string
	Kind         string
	APIVersion   string
	Name         string
	Synced       string
	SyncedReason string
	Ready        string
	ReadyReason  string
}

func (r Resource) ToRow() []string {
	return []string{
		r.Namespace,
		r.Kind,
		r.APIVersion,
		r.Name,
		r.Synced,
		r.SyncedReason,
		r.Ready,
		r.ReadyReason,
	}
}

func Headers() []string {
	return []string{
		"Namespace",
		"Kind",
		"APIVersion",
		"Name",
		"Synced",
		"SyncedReason",
		"Ready",
		"ReadyReason",
	}
}
