package ui

import (
	"fmt"
	"strings"

	"github.com/nkzk/xtree/internal/utils"
)

type Client interface {
	GetXR() string
	Get(kind, apiversion, name, namespace string) string
}

type mock struct{}

type kubectl struct{}

func (k kubectl) GetXR(kind, apiversion, name, namespace string) (string, error) {
	output, err := utils.RunCommand("kubectl", "get", apiversion+"/"+name, "-n", namespace, "-o", "yaml")
	if err != nil {
		return "", fmt.Errorf("failed to get XR: %w", err)
	}
	return string(output), err
}

func (m mock) GetXR(kind, apiversion, name, namespace string) (string, error) {
	return strings.TrimPrefix(`
apiVersion: group.domain.com/v1alpha1
kind: Auth
metadata:
  labels:
    crossplane.io/composite: example
  name: example
  namespace: default
  resourceVersion: "803501"
  uid: 76102637-8913-4ef0-8205-dff9bf75ef04
spec:
  crossplane:
    compositionRef:
      name: kinds.group.domain.com
    compositionRevisionRef:
      name: kinds.group.domain.com
    compositionUpdatePolicy: Automatic
    resourceRefs:
    - apiVersion: app.azuread.m.upbound.io/v1beta1
      kind: RoleAssignment
      name: objectid-app-assignment-0
    - apiVersion: applications.azuread.m.upbound.io/v1beta1
      kind: Application
      name: example-app
    - apiVersion: applications.azuread.m.upbound.io/v1beta1
      kind: Password
      name: example-password
    - apiVersion: policies.azuread.m.upbound.io/v1beta1
      kind: ClaimsMappingPolicy
      name: samaccountname-policy
    - apiVersion: protection.crossplane.io/v1beta1
      kind: Usage
      name: example-password-usage
    - apiVersion: serviceprincipaldelegated.azuread.m.upbound.io/v1beta1
      kind: PermissionGrant
      name: grant
    - apiVersion: serviceprincipals.azuread.m.upbound.io/v1beta1
      kind: ClaimsMappingPolicyAssignment
      name: samaccountname-policy-assignment
    - apiVersion: serviceprincipals.azuread.m.upbound.io/v1beta1
      kind: Principal
      name: example-sp
    - apiVersion: serviceprincipals.azuread.m.upbound.io/v1beta1
      kind: Principal
      name: graph-principal
    - apiVersion: v1
      kind: Secret
      name: example-connection-details
  enableSamAccountNameMapping: true
  groupMembershipClaims:
  - SecurityGroup
  optionalClaims:
    idToken:
    - essential: false
      name: email
    - essential: false
      name: upn
  publicClient:
    redirectUris:
    - http://localhost:3000/callback
  signInAudience: AzureADMyOrg
  web:
    redirectUris:
    - https://example.x.no/callback
    - https://example.x.no/callback2
status:
  clientId: 1b897907-0d81-44f7-a7bf-549e72abe17a
  conditions:
  - lastTransitionTime: "2025-10-10T12:55:42Z"
    observedGeneration: 7
    reason: ReconcileSuccess
    status: "True"
    type: Synced
  - lastTransitionTime: "2025-10-10T12:55:42Z"
    message: 'Unready resources: example-connection-details'
    observedGeneration: 7
    reason: Creating
    status: "False"
    type: Ready
  graphObjectId: x
  principalObjectId: x
  tenantId: x
`,
		"\n"), nil
}

func (m mock) Get(kind, apiversion, name, namespace string) string {
	return fmt.Sprintf("ayo %s", name)
}
