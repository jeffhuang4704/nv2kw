## Group policy conversion experiment

### The experiement
TODO: describe Test1 and Test2
and the confusion of it the && , ||

### Test 1 - reject_latest() && use_ban_label()

A group policy consists of two sub-policies defined by the expression: reject_latest() && use_ban_label().

<details><summary>yaml</summary>

```
neuvector@ubuntu2204-F:~/kubewarden/test$ cat  grouppolicy1.yaml
apiVersion: policies.kubewarden.io/v1
kind: ClusterAdmissionPolicyGroup # or AdmissionPolicyGroup
metadata:
  name: demo1
spec:
  rules:
    - apiGroups: ["*"]
      apiVersions: ["*"]
      resources: ["*"]
      operations:
        - CREATE
        - UPDATE
  policies:
    use_ban_label:
      module: ghcr.io/kubewarden/policies/safe-labels:v0.1.14
      settings:
        denied_labels:
          - ban1
          - ban2
    reject_latest:
      module: ghcr.io/kubewarden/policies/trusted-repos:v0.2.0
      settings:
        tags:
          reject:
            - latest
  expression: "reject_latest() && use_ban_label()"        👈
  message: "rejected - reject_latest() && use_ban_label()"

```

```
neuvector@ubuntu2204-F:~/kubewarden/test$ kubectl get capg
NAME    POLICY SERVER   MUTATING   BACKGROUNDAUDIT   MODE      OBSERVED MODE   STATUS   AGE
demo1   default                    true              protect   protect         active   3m23s
```
</details>

Resource 1 uses the `ban1` label, and its evaluation result is rejected.

<details><summary>yaml and evaluation result</summary>
```
neuvector@ubuntu2204-F:~/kubewarden/test$ cat 1_deploy-label.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: my-dep
    ban1: ttt      👈
  name: my-dep
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-dep
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: my-dep
    spec:
      containers:
      - image: busybox:v1
        name: busybox
        resources: {}
status: {}
```

```
## rejected ❌
neuvector@ubuntu2204-F:~/kubewarden/test$ kubectl apply -f 1_deploy-label.yaml
Error from server: error when creating "1_deploy-label.yaml": admission webhook "clusterwide-group-demo1.kubewarden.admission" denied the request: rejected - reject_latest() && use_ban_label()
```
</details>


Resource 2 uses `latest` tag
```
neuvector@ubuntu2204-F:~/kubewarden/test$ cat 2_deploy-latest.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: my-dep
  name: my-dep
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-dep
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: my-dep
    spec:
      containers:
      - image: busybox:latest   👈
        name: busybox
        resources: {}
status: {}

## rejected ❌
neuvector@ubuntu2204-F:~/kubewarden/test$ kubectl apply -f 2_deploy-latest.yaml
Error from server: error when creating "2_deploy-latest.yaml": admission webhook "clusterwide-group-demo1.kubewarden.admission" denied the request: rejected - reject_latest() && use_ban_label()
```

Resource 3 uses `latest` tag and `ban1` label
```
neuvector@ubuntu2204-F:~/kubewarden/test$ cat 3_deploy-latest_and_banned_label.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: my-dep
    ban1: tty
  name: my-dep
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-dep
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: my-dep
    spec:
      containers:
      - image: busybox:latest
        name: busybox
        resources: {}
status: {}

## rejected ❌
neuvector@ubuntu2204-F:~/kubewarden/test$ kubectl apply -f 3_deploy-latest_and_banned_label.yaml
Error from server: error when creating "3_deploy-latest_and_banned_label.yaml": admission webhook "clusterwide-group-demo1.kubewarden.admission" denied the request: rejected - reject_latest() && use_ban_label()
```

| Resource    | Result                                    |
| ---------- | ---------------------------------------------- |
| Resource 1 uses `ban1` label | rejected ❌              |
| Resource 2 uses `latest` tag  | rejected ❌ |
| Resource 3 uses `latest` tag and `ban1` label |rejected ❌ |

### Test 2 - reject_latest() || use_ban_label()

The ClusterAdmissionPolicyGroup remains the same, except the expression has been changed from && to ||.

```
neuvector@ubuntu2204-F:~/kubewarden/test$ cat grouppolicy2.yaml
apiVersion: policies.kubewarden.io/v1
kind: ClusterAdmissionPolicyGroup # or AdmissionPolicyGroup
metadata:
  name: demo2
spec:
  rules:
    - apiGroups: ["*"]
      apiVersions: ["*"]
      resources: ["*"]
      operations:
        - CREATE
        - UPDATE
  policies:
    use_ban_label:
      module: ghcr.io/kubewarden/policies/safe-labels:v0.1.14
      settings:
        denied_labels:
          - ban1
          - ban2
    reject_latest:
      module: ghcr.io/kubewarden/policies/trusted-repos:v0.2.0
      settings:
        tags:
          reject:
            - latest
  expression: "reject_latest() || use_ban_label()"   👈 
  message: "rejected - reject_latest() || use_ban_label()"
```



Resource 1 uses `ban1` label
```
neuvector@ubuntu2204-F:~/kubewarden/test$ cat 1_deploy-label.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: my-dep
    ban1: ttt      👈
  name: my-dep
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-dep
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: my-dep
    spec:
      containers:
      - image: busybox:v1
        name: busybox
        resources: {}
status: {}

# allowed  ✔️
neuvector@ubuntu2204-F:~/kubewarden/test$ kubectl apply -f 1_deploy-label.yaml
deployment.apps/my-dep created
```


Resource 2 uses `latest` tag
```
neuvector@ubuntu2204-F:~/kubewarden/test$ cat 2_deploy-latest.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: my-dep
  name: my-dep
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-dep
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: my-dep
    spec:
      containers:
      - image: busybox:latest   👈
        name: busybox
        resources: {}
status: {}

# allowed  ✔️
neuvector@ubuntu2204-F:~/kubewarden/test$ kubectl apply -f 2_deploy-latest.yaml
deployment.apps/my-dep created
```


Resource 3 uses `latest` tag and `ban1` label
```
neuvector@ubuntu2204-F:~/kubewarden/test$ cat 3_deploy-latest_and_banned_label.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: my-dep
    ban1: tty
  name: my-dep
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-dep
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: my-dep
    spec:
      containers:
      - image: busybox:latest
        name: busybox
        resources: {}
status: {}

## rejected ❌
neuvector@ubuntu2204-F:~/kubewarden/test$ kubectl apply -f 3_deploy-latest_and_banned_label.yaml
Error from server: error when creating "3_deploy-latest_and_banned_label.yaml": admission webhook "clusterwide-group-demo2.kubewarden.admission" denied the request: rejected - reject_latest() || use_ban_label()
```

### The Doc

```
Expression is the evaluation expression to accept or reject the admission request under evaluation. This field uses CEL as the expression language for the policy groups. Each policy in the group will be represented as a function call in the expression with the same name as the policy defined in the group. The expression field should be a valid CEL expression that evaluates to a boolean value. If the expression evaluates to true, the group policy will be considered as accepted, otherwise, it will be considered as rejected. This expression allows grouping policies calls and perform logical operations on the results of the policies. See Kubewarden documentation to learn about all the features available.
```