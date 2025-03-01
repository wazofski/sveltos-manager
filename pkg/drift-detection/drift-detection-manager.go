// Generated by *go generate* - DO NOT EDIT
/*
Copyright 2023. projectsveltos.io. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package driftdetection

var driftDetectionYAML = []byte(`apiVersion: v1
kind: Namespace
metadata:
  labels:
    control-plane: drift-detection-manager
  name: projectsveltos
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: drift-detection-manager
  namespace: projectsveltos
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: drift-detection-leader-election-role
  namespace: projectsveltos
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: drift-detection-manager-role
rules:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - lib.projectsveltos.io
  resources:
  - debuggingconfigurations
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - lib.projectsveltos.io
  resources:
  - resourcesummaries
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - lib.projectsveltos.io
  resources:
  - resourcesummaries/finalizers
  verbs:
  - update
- apiGroups:
  - lib.projectsveltos.io
  resources:
  - resourcesummaries/status
  verbs:
  - get
  - patch
  - update
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: drift-detection-metrics-reader
rules:
- nonResourceURLs:
  - /metrics
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: drift-detection-proxy-role
rules:
- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create
- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: drift-detection-leader-election-rolebinding
  namespace: projectsveltos
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: drift-detection-leader-election-role
subjects:
- kind: ServiceAccount
  name: drift-detection-manager
  namespace: projectsveltos
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: drift-detection-manager-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: drift-detection-manager-role
subjects:
- kind: ServiceAccount
  name: drift-detection-manager
  namespace: projectsveltos
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: drift-detection-proxy-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: drift-detection-proxy-role
subjects:
- kind: ServiceAccount
  name: drift-detection-manager
  namespace: projectsveltos
---
apiVersion: v1
data:
  controller_manager_config.yaml: |
    apiVersion: controller-runtime.sigs.k8s.io/v1alpha1
    kind: ControllerManagerConfig
    health:
      healthProbeBindAddress: :8081
    metrics:
      bindAddress: 127.0.0.1:8080
    webhook:
      port: 9443
    leaderElection:
      leaderElect: true
      resourceName: 563581df.projectsveltos.io
    # leaderElectionReleaseOnCancel defines if the leader should step down volume
    # when the Manager ends. This requires the binary to immediately end when the
    # Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
    # speeds up voluntary leader transitions as the new leader don't have to wait
    # LeaseDuration time first.
    # In the default scaffold provided, the program ends immediately after
    # the manager stops, so would be fine to enable this option. However,
    # if you are doing or is intended to do any operation such as perform cleanups
    # after the manager stops then its usage might be unsafe.
    # leaderElectionReleaseOnCancel: true
kind: ConfigMap
metadata:
  name: drift-detection-manager-config
  namespace: projectsveltos
---
apiVersion: v1
kind: Service
metadata:
  labels:
    control-plane: drift-detection-manager
  name: drift-detection-manager-metrics-service
  namespace: projectsveltos
spec:
  ports:
  - name: https
    port: 8443
    protocol: TCP
    targetPort: https
  selector:
    control-plane: drift-detection-manager
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    control-plane: drift-detection-manager
  name: drift-detection-manager
  namespace: projectsveltos
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: drift-detection-manager
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/default-container: manager
      labels:
        control-plane: drift-detection-manager
    spec:
      containers:
      - args:
        - --health-probe-bind-address=:8081
        - --metrics-bind-address=127.0.0.1:8080
        - --leader-elect
        - --v=5
        - --cluster-namespace=
        - --cluster-name=
        - --cluster-type=
        - --run-mode=do-not-send-updates
        command:
        - /manager
        image: gianlucam76/drift-detection-manager-amd64:main
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 15
          periodSeconds: 20
        name: manager
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          limits:
            cpu: 500m
            memory: 128Mi
          requests:
            cpu: 10m
            memory: 64Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
      - args:
        - --secure-listen-address=0.0.0.0:8443
        - --upstream=http://127.0.0.1:8080/
        - --logtostderr=true
        - --v=0
        image: gcr.io/kubebuilder/kube-rbac-proxy:v0.12.0
        name: kube-rbac-proxy
        ports:
        - containerPort: 8443
          name: https
          protocol: TCP
        resources:
          limits:
            cpu: 500m
            memory: 128Mi
          requests:
            cpu: 5m
            memory: 64Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
      securityContext:
        runAsNonRoot: true
      serviceAccountName: drift-detection-manager
      terminationGracePeriodSeconds: 10
`)
