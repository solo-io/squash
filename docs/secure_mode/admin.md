# Secure Mode Administration

# Introduction
- Secure Mode is designed to allow you to apply regular Kubernetes RBAC configurations to Squash debugging activities.
- **Proper configuration is required for Secure Mode to be secure**.
- Fortunately, configuration is easy. This guide outlines the best practices for configuring Secure Mode. Ultimately, you are responsible for your cluster's security, so please apply these recommendations as is appropriate to your use case.

# Motivation
- Secure Mode is recommended for multi-user clusters.
- Reasons to use Secure Mode:
  - Prevent users from unexpectedly halting your process in a debug state.
  - Prevent malicious exploitation of the Plank pods' `SYS_PTRACE` capabilities.
  - Prevent malicious exploitation of the Squash pod's `Privileged` security context.

# Deploy Squash
## Quick Start
- To "kick the tires" on Squash's Secure Mode, you can install it with a single command: `squashctl deploy squash`.
- This creates all the resources needed to start using Secure Mode:
  - Creates `squash-debugger` namespace.
  - Creates a Service Account with the minimal required permissions.
  - Deploys Squash.
## Formal Deployment
- For shared cluster Squash usage, you should manage your squash deployment through your conventional workflow.
- Resources required by Squash include:
  - Deployments - Squash
  - Service Accounts - Squash, Plank
  - Cluster Roles - Squash, Plank
  - Cluster Role Bindings - Squash, Plank
- For details, see a reference configuration for these resources below.

# Configuration requirements for preventing undesired debug activities
- Suggestion: **do not** authorize your users to `kubectl exec` into the namespace that stores the `squash` and `plank` pods.
```yaml
# Reference: pod exec Policy Rule
# DO NOT enable this permission for the squash-debugger namespace
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: default # DO NOT enable this permission for the squash-debugger namespace
  name: pod-exec
rules:
- apiGroups: [""]
  resources: ["pods/exec"] # DO NOT enable this permission for the squash-debugger namespace
  verbs: ["create"]
  ```

# Reference configuration

- Note: This is a reference configuration for the resources required by Squash.
- Note: A priority for our implementation of Secure Mode is minimizing permissions granted to Squash and Plank pods. Our roadmap includes upcoming features that will further reduce the permissions needed by Squash and Plank. These future changes are cited below. We will update these reference configurations as their needs change.

## Managed resources

- These are the resources you need to configure and manage yourself.

### Deployment - Squash
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  annotations:
    deployment.kubernetes.io/revision: "1"
  labels:
    app: squash
  name: squash
  namespace: squash-debugger
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: squash
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: squash
    spec:
      containers:
      - env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: HOST_ADDR
          value: $(POD_NAME).$(POD_NAMESPACE)
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        image: soloio/squash:dev
        imagePullPolicy: IfNotPresent
        name: squash
        ports:
        - containerPort: 1234
          name: http
          protocol: TCP
        resources: {}
        securityContext:
          privileged: true
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/run/cri.sock
          name: crisock
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: squash
      serviceAccountName: squash
      terminationGracePeriodSeconds: 30
      volumes:
      - hostPath:
          path: /var/run/dockershim.sock
          type: ""
        name: crisock
```

### Service Account - Squash
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: squash
  namespace: squash-debugger
```

### Cluster Role Binding - Squash
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: squash-crb-pods
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: squash-cr-pods
subjects:
- kind: ServiceAccount
  name: squash
  namespace: squash-debugger
```

### Cluster Role - Squash
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: squash-cr-pods
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
  - create
  - delete
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - list
- apiGroups:
  - squash.solo.io
  resources:
  - debugattachments
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - clusterrolebindings
  verbs:
  - create
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - clusterrole
  verbs:
  - create
- apiGroups:
  - ""
  resources:
# This rule can be removed if you pre-allocate service accounts for Plank
# This rule not be needed in future versions of Squash
# when Squash will expect Plank service accounts already to exist
  - serviceaccount
  verbs:
  - create
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
  - register
```

### Service Account - Plank
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: squash-plank
  namespace: squash-debugger
```

### Cluster Role Binding - Plank
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: squash-plank-crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: squash-plank-cr
subjects:
- kind: ServiceAccount
  name: squash-plank
  namespace: squash-debugger
```

### Cluster Role - Plank
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: squash-plank-cr
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
  - create
  - delete
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - list
- apiGroups:
  - squash.solo.io
  resources:
  - debugattachments
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
  - register # This will not be needed in future versions
```

## Implicit Resources
- These resources are managed by Squash itself.
- Note: Squash currently has the ability to create the Service Account, Cluster Role, and Cluster Role Bindings needed by Plank pods. As noted above, this capability will be revoked in the near-term and Squash will expect them to have already been created.
### Pod - Plank
- Note: Plank pods are created by Squash, you will not need to manage this resource but it is useful to know what it looks like.
- In this example, the is debugging a pod named "example-service1-8499d97885" in namespace "my-namespace"
```yaml
apiVersion: v1
kind: Pod
metadata:
  generateName: plank
  labels:
    debug_attachment_name: example-service1-8499d97885-bgh59
    debug_attachment_namespace: my-namespace
    squash: plank
  name: planklrv4j
  namespace: squash-debugger
spec:
  containers:
  - env:
    - name: SQUASH_DEBUG_ATTACHMENT_NAMESPACE
      value: my-namespace
    - name: SQUASH_DEBUG_ATTACHMENT_NAME
      value: example-service1-8499d97885-bgh59-example-service1
    image: soloio/plank-dlv:dev
    imagePullPolicy: IfNotPresent
    name: plank
    resources: {}
    securityContext:
      capabilities:
        add:
        - SYS_PTRACE
      procMount: Default
    stdin: true
    stdinOnce: true
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    tty: true
    volumeMounts:
    - mountPath: /var/run/cri.sock
      name: crisock
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: squash-plank-token-jjjxq
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  hostPID: true
  nodeName: minikube
  priority: 0
  restartPolicy: Never
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: squash-plank
  serviceAccountName: squash-plank
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  volumes:
  - hostPath:
      path: /var/run/dockershim.sock
      type: ""
    name: crisock
  - name: squash-plank-token-jjjxq
    secret:
      defaultMode: 420
      secretName: squash-plank-token-jjjxq
```
