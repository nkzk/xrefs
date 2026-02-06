package ui

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type resourceRef struct {
	ApiVersion string `json:"apiVersion" yaml:"apiVersion"`
	Kind       string `json:"kind" yaml:"kind"`
	Name       string `json:"name" yaml:"name"`
}

type XR struct {
	Metadata struct {
		Namespace string `json:"namespace" yaml:"namespace"`
	} `json:"metadata" yaml:"metadata"`
	Spec struct {
		Crossplane struct {
			ResourceRefs []resourceRef `json:"resourceRefs" yaml:"resourceRefs"`
		} `json:"crossplane" yaml:"crossplane"`
	} `json:"spec" yaml:"spec"`
}

type condition struct {
	Status        string `json:"status" yaml:"status"`
	ConditionType string `json:"type" yaml:"type"`
	Reason        string `json:"reason" yaml:"reason"`
}

type status struct {
	Conditions conditions `json:"conditions" yaml:"conditions"`
}

type conditions []condition

//	Get a condition
//
// Example usage:
//
//	conditions.Get("Ready").Status
//	conditions.Get("Synced").Status
func (c conditions) Get(s string) condition {
	for _, condition := range c {
		if condition.ConditionType == s {
			return condition
		}
	}

	return condition{}
}
