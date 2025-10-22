package ui

import (
	"fmt"
	"strings"

	"github.com/nkzk/xrefs/internal/utils"
)

type Client interface {
	GetXR(command string) (string, error)
	Get(command string) (string, error)
}

type mock struct{}

type kubectl struct{}

func NewKubectlClient() *kubectl {
	return &kubectl{}
}

func NewMockClient() *mock {
	return &mock{}
}

func createGetYamlCommand(kind, group, apiversion, name, namespace string) (string, error) {
	if kind == "" || apiversion == "" || name == "" || namespace == "" {
		return "", fmt.Errorf("missing kind, apiversion, name or namespace")
	}

	// if apiVersion contains "/" it means it also contains the group, return accordingly
	i := strings.IndexRune(apiversion, '/')
	if i != -1 {
		apiversion = apiversion[:i]
		return fmt.Sprintf("kubectl get %s.%s/%s -n %s -o yaml", kind, apiversion, name, namespace), nil
	}

	return fmt.Sprintf("kubectl get %s.%s.%s/%s -n %s -o yaml", kind, apiversion, group, name, namespace), nil
}

func createDescribeCommand(row row) (string, error) {
	if row.Kind == "" || row.ApiVersion == "" || row.Name == "" || row.Namespace == "" {
		return "", fmt.Errorf("missing kind, apiversion, name or namespace")
	}

	apiVersion := row.ApiVersion
	// if apiVersion contains "/" it means it also contains the group, return accordingly
	i := strings.IndexRune(apiVersion, '/')
	if i != -1 {
		apiVersion = apiVersion[:i]
		return fmt.Sprintf("kubectl describe %s.%s/%s -n %s", row.Kind, row.ApiVersion, row.Name, row.Namespace), nil
	}

	return fmt.Sprintf("kubectl describe %s.%s.%s/%s -n %s", row.Kind, row.ApiVersion, row.Name, row.Namespace), nil
}

func (k kubectl) GetXR(command string) (string, error) {
	cmd := strings.Fields(command)

	output, err := utils.RunCommand(cmd[0], cmd[1:]...)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %s, %w", cmd, err)
	}

	return string(output), err
}

func (m mock) GetXR(command string) (string, error) {
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

func (m mock) Get(command string) (string, error) {
	// if describe
	if strings.Contains(command, "describe") {
		return fmt.Sprintf(`
Name:                 coredns-5d78c9869d-dxvz8
Namespace:            kube-system
Priority:             2000000000
Priority Class Name:  system-cluster-critical
Service Account:      coredns
Node:                 kind-control-plane/172.18.0.2
Start Time:           Sun, 05 Oct 2025 16:03:16 +0200
Labels:               k8s-app=kube-dns
                      pod-template-hash=5d78c9869d
Annotations:          <none>
Status:               Running
IP:                   10.244.0.3
IPs:
  IP:           10.244.0.3
Controlled By:  ReplicaSet/coredns-5d78c9869d
Containers:
  coredns:
    Container ID:  containerd://3f83bb8b96d18496801e6d607dde71e3369cff79f56b45dbea220ba6e32fc54a
    Image:         registry.k8s.io/coredns/coredns:v1.10.1
    Image ID:      sha256:97e04611ad43405a2e5863ae17c6f1bc9181bdefdaa78627c432ef754a4eb108
    Ports:         53/UDP (dns), 53/TCP (dns-tcp), 9153/TCP (metrics)
    Host Ports:    0/UDP (dns), 0/TCP (dns-tcp), 0/TCP (metrics)
    Args:
      -conf
      /etc/coredns/Corefile
    State:          Running
      Started:      Sun, 05 Oct 2025 16:03:18 +0200
    Ready:          True
    Restart Count:  0
    Limits:
      memory:  170Mi
    Requests:
      cpu:        100m
      memory:     70Mi
    Liveness:     http-get http://:8080/health delay=60s timeout=5s period=10s #success=1 #failure=5
    Readiness:    http-get http://:8181/ready delay=0s timeout=1s period=10s #success=1 #failure=3
    Environment:  <none>
    Mounts:
      /etc/coredns from config-volume (ro)
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-hp52t (ro)
Conditions:
  Type              Status
  Initialized       True 
  Ready             True 
  ContainersReady   True 
  PodScheduled      True 
Volumes:
  config-volume:
    Type:      ConfigMap (a volume populated by a ConfigMap)
    Name:      coredns
    Optional:  false
  kube-api-access-hp52t:
    Type:                    Projected (a volume that contains injected data from multiple sources)
    TokenExpirationSeconds:  3607
    ConfigMapName:           kube-root-ca.crt
    Optional:                false
    DownwardAPI:             true
QoS Class:                   Burstable
Node-Selectors:              kubernetes.io/os=linux
Tolerations:                 CriticalAddonsOnly op=Exists
                             node-role.kubernetes.io/control-plane:NoSchedule
                             node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                             node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:                      <none>
`), nil
	}

	// if yaml
	return fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2025-10-05T14:03:13Z"
  generateName: coredns-5d78c9869d-
  labels:
    k8s-app: kube-dns
    pod-template-hash: 5d78c9869d
  name: coredns-5d78c9869d-dxvz8
  namespace: kube-system
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: coredns-5d78c9869d
    uid: 002a4486-a828-4927-b7ee-eb0bce4b47f0
  resourceVersion: "430"
  uid: 53bb2cbd-2e51-431f-b1fa-ee9ddcb018dc
spec:
  affinity:
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
      - podAffinityTerm:
          labelSelector:
            matchExpressions:
            - key: k8s-app
              operator: In
              values:
              - kube-dns
          topologyKey: kubernetes.io/hostname
        weight: 100
  containers:
  - args:
    - -conf
    - /etc/coredns/Corefile
    image: registry.k8s.io/coredns/coredns:v1.10.1
    imagePullPolicy: IfNotPresent
    livenessProbe:
      failureThreshold: 5
      httpGet:
        path: /health
        port: 8080
        scheme: HTTP
      initialDelaySeconds: 60
      periodSeconds: 10
      successThreshold: 1
      timeoutSeconds: 5
    name: coredns
    ports:
    - containerPort: 53
      name: dns
      protocol: UDP
    - containerPort: 53
      name: dns-tcp
      protocol: TCP
    - containerPort: 9153
      name: metrics
      protocol: TCP
    readinessProbe:
      failureThreshold: 3
      httpGet:
        path: /ready
        port: 8181
        scheme: HTTP
      periodSeconds: 10
      successThreshold: 1
      timeoutSeconds: 1
    resources:
      limits:
        memory: 170Mi
      requests:
        cpu: 100m
        memory: 70Mi
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        add:
        - NET_BIND_SERVICE
        drop:
        - all
      readOnlyRootFilesystem: true
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /etc/coredns
      name: config-volume
      readOnly: true
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: kube-api-access-hp52t
      readOnly: true
  dnsPolicy: Default
  enableServiceLinks: true
  nodeName: kind-control-plane
  nodeSelector:
    kubernetes.io/os: linux
  preemptionPolicy: PreemptLowerPriority
  priority: 2000000000
  priorityClassName: system-cluster-critical
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: coredns
  serviceAccountName: coredns
  terminationGracePeriodSeconds: 30
  tolerations:
  - key: CriticalAddonsOnly
    operator: Exists
  - effect: NoSchedule
    key: node-role.kubernetes.io/control-plane
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  volumes:
  - configMap:
      defaultMode: 420
      items:
      - key: Corefile
        path: Corefile
      name: coredns
    name: config-volume
  - name: kube-api-access-hp52t
    projected:
      defaultMode: 420
      sources:
      - serviceAccountToken:
          expirationSeconds: 3607
          path: token
      - configMap:
          items:
          - key: ca.crt
            path: ca.crt
          name: kube-root-ca.crt
      - downwardAPI:
          items:
          - fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
            path: namespace
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2025-10-05T14:03:16Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2025-10-05T14:03:18Z"
    status: "True"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2025-10-05T14:03:18Z"
    status: "True"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2025-10-05T14:03:16Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - containerID: containerd://3f83bb8b96d18496801e6d607dde71e3369cff79f56b45dbea220ba6e32fc54a
    image: registry.k8s.io/coredns/coredns:v1.10.1
    imageID: sha256:97e04611ad43405a2e5863ae17c6f1bc9181bdefdaa78627c432ef754a4eb108
    lastState: {}
    name: coredns
    ready: true
    restartCount: 0
    started: true
    state:
      running:
        startedAt: "2025-10-05T14:03:18Z"
  hostIP: 172.18.0.2
  phase: Running
  podIP: 10.244.0.3
  podIPs:
  - ip: 10.244.0.3
  qosClass: Burstable
  startTime: "2025-10-05T14:03:16Z"
`), nil
}

func (k kubectl) Get(command string) (string, error) {
	cmd := strings.Fields(command)

	output, err := utils.RunCommand(cmd[0], cmd[1:]...)
	if err != nil {
		return "", fmt.Errorf("failed to get resource: %v", err)
	}

	return string(output), nil
}
