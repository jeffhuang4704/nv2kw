## Enivronment variables

### Background

According to the example in [environment-variable-policy doc](https://github.com/kubewarden/environment-variable-policy?tab=readme-ov-file#examples
In the following example, only resources that have the `envvar3` or `envvar2` defined will be allowed:
```
settings:
  rules:
    - reject: anyNotIn
      environmentVariables:
        - name: "envvar2"
          value: "envvar2_value"
        - name: "envvar3"
```

Results of an experiment indicate a resource must use both `envvar3` and `envvar2` in order to be allowed.
| Resource    | Result                                    |
| ---------- | ---------------------------------------------- |
| Resource 1 uses `envvar2` and `envvar3` label | accepted ✔️              |
| Resource 2 uses `envvar2`   | rejected ❌ |
| Resource 3 uses `envvar3` |rejected ❌ |



<details><summary>Kubewarden Policy using anyNotIn</summary>

```
neuvector@ubuntu2204-F:~/kubewarden/test_env$ kubectl get cap
NAME   POLICY SERVER   MUTATING   BACKGROUNDAUDIT   MODE      OBSERVED MODE   STATUS   AGE
env1   default         false      true              protect   protect         active   8m18s

neuvector@ubuntu2204-F:~/kubewarden/test_env$ kubectl get cap env1 -oyaml
apiVersion: policies.kubewarden.io/v1
kind: ClusterAdmissionPolicy
metadata:
  ....
spec:
  backgroundAudit: true
  mode: protect
  module: ghcr.io/kubewarden/policies/environment-variable-policy:v0.1.7
  mutating: false
  policyServer: default
  rules:
     ...
  settings:
    rules:
    - environmentVariables:
      - name: envvar2
        value: envvar2_value
      - name: envvar3
      reject: anyNotIn
  timeoutSeconds: 10
```
</details>

The following are the resources and their execution results:

<details><summary>Resource 1 uses both `envvar2` and `envvar3` environment variables, and its evaluation result is allowed.</summary>

```
# resource using both envvar2 and envvar3
neuvector@ubuntu2204-F:~/kubewarden/test_env$ cat 1_deploy-env.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
    ...
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
        env:
        - name: envvar2
          value: envvar2_value
        - name: envvar3
          value: aaaaaa
status: {}

# deploy -> allowed
neuvector@ubuntu2204-F:~/kubewarden/test_env$ kubectl apply -f 1_deploy-env.yaml
deployment.apps/my-dep created
```
</details>


<details><summary>Resource 2 uses only `envvar2` environment variable, and its evaluation result is denied.</summary>

```
# resource using envvar2
neuvector@ubuntu2204-F:~/kubewarden/test_env$ cat 1_deploy-env.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
    ...
spec:
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
        env:
        - name: envvar2
          value: envvar2_value
status: {}

# deploy -> denied
neuvector@ubuntu2204-F:~/kubewarden/test_env$ kubectl apply -f 1_deploy-env.yaml
Error from server: error when creating "1_deploy-env.yaml": admission webhook "clusterwide-env1.kubewarden.admission" denied the request: Resource should define at least one of the environment variables from the rule. Invalid environment variables found: envvar3
```
</details>

<details><summary>Resource 3 uses only `envvar3` environment variable, and its evaluation result is denied.</summary>

```
# resource using envvar3
neuvector@ubuntu2204-F:~/kubewarden/test_env$ cat 1_deploy-env.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
    ...
spec:
    ...
  template:
    spec:
      containers:
      - image: nginx
        name: nginx
        env:
        - name: envvar3
          value: envvar3_value
status: {}

# deploy -> denied
neuvector@ubuntu2204-F:~/kubewarden/test_env$ kubectl apply -f 1_deploy-env.yaml
Error from server: error when creating "1_deploy-env.yaml": admission webhook "clusterwide-env1.kubewarden.admission" denied the request: Resource should define at least one of the environment variables from the rule. Invalid environment variables found: envvar2
```
</details>
