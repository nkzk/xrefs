package ui

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestGetStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   *status
	}{
		{
			name: "Ready true",
			status: strings.TrimPrefix(`
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2025-10-05T14:03:18Z"
    status: "True"
    type: Ready`, "\n"),
			want: &status{
				Conditions: []condition{
					{
						Status:        "True",
						ConditionType: "Ready",
						Reason:        "",
					},
				},
			},
		},
		{
			name: "Empty conditions",
			status: strings.TrimPrefix(`
status: {}`, "\n"),
			want: &status{
				Conditions: []condition{},
			},
		},
		{
			name: "Synced and Ready true",
			status: strings.TrimPrefix(`
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2025-10-05T14:03:18Z"
    reason: "prokrustesseng"
    status: "True"
    type: Synced
  - lastProbeTime: null
    lastTransitionTime: "2025-10-05T14:03:18Z"
    status: "True"
    reason: "gullfeber"
    type: Ready`, "\n"),
			want: &status{
				Conditions: []condition{
					{
						Status:        "True",
						ConditionType: "Synced",
						Reason:        "prokrustesseng",
					},
					{
						Status:        "True",
						ConditionType: "Ready",
						Reason:        "gullfeber",
					},
				},
			},
		},
	}

	for _, test := range tests {
		status, err := getStatus(test.status)
		if err != nil {
			t.Fatalf("failed to get status: %v", err)
		}

		if equal := reflect.DeepEqual(status, test.want); !equal {
			prettyFatal(t, *test.want, *status)
		}
	}

}

func prettyFatal(t *testing.T, want, got status) {
	var err error
	var w, g []byte
	w, err = json.MarshalIndent(want, "", "  ")
	g, err = json.MarshalIndent(got, "", "  ")

	if err != nil {
		t.Fatalf("want: %+v got: %+v", want, got)
	}

	t.Fatalf("unexpected result:\n\033[1m\033[32mwant:\033[0m\n%s\n\033[1m\033[31mgot:\033[0m\n%s", w, g)
}
