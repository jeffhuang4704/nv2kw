## cel policy notes

## resources
-- [cel language def](https://github.com/google/cel-spec/blob/v0.18.0/doc/langdef.md#macros)
-- [CEL playground](https://playcel.undistro.io/)

## notes
We can leverage the Kubewarden CEL policy to implement the following NeuVector criteria.
- labels
- annotation
- environment variables

in Kubewarden CEL policy => The policy will be evaluated as allowed if all the CEL expressions are evaluated as true

The following operators are used in these criteria.

```
"containsAll"
    The resource will be denied if it uses ALL of the defined values.
    The resource will not be denied if it uses some or none of the defined values.

"containsAny"
    The resource will be denied if it uses any of the defined values.

"notContainsAny"
    It is the negation of "containsAny".
    The resource will be denied if it does not use any of the defined values.
    In other words, the resource must use at least one of the defined values to be allowed.

"containsOtherThan"
    The resource will be denied if its usage does not match any of the defined values.
    In other words, the defined values act as a whitelist, and any usage outside the whitelist will be denied.
    The resource may only use the defined values (whitelist).
```

The following are some experiments attempting to implement these criteria and their operators. 
For each, we need to consider the following scenarios, using labels as an example.

-- scenario 1 - only label key is used  
-- scenario 2 - label key and value  
-- scenario 3 - label key and value, and the value has regex  

🚧 TODO: need to mention the object path each resource wll inspect. For example label it might inspect two different places.
Also the environment variables, it might inspect both image and runtime..



<details><summary>Example Input (use this in CEL playground)</summary>

```
params:
  allowedRegistries: 
    - myregistry.com
    - docker.io # use 'docker.io' for Docker Hub
object:
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    template:
      metadata:
        name: nginx
        labels:
          app: nginx
          badlabel1: badvalue1
          badlabel2: aa
          badlabel3: bb
      spec:
        containers:
          - name: nginx
            image: nginx # the expression looks for this field
    selector:
      matchLabels:
        app: nginx
```

</details>

### Label criteria and value types

* 🔴 operator = containsAll
* 🔴 operator = containsAny
* 🔴 operator = notContainsAny
* 🔴 operator = containsOtherThan

For each operator, the following scenarios should be considered.

* scenario 1 -  label key is used, example: `value = ["badlabel1","badlabel2","badlabel3"]`
* scenario 2a - label key and value, example: `value = ["badlabel1=badvalue1"]`  
* scenario 2b - multiple label key and value, example: `value = ["badlabel1=badvalue1", "badlabel2=badvalue2"]`  
* scenario 3a - label key and regex value, example: `value = ["badlabel1=^bad*"]`  
* scenario 3b - multiple label key and regex value example: `value = ["badlabel1=^bad*", "badlabel2=^anotherbad*"] ` 
* scenario 4  - mixed type of values = `["badlabel1", "badlabel2=badvalue2"]`  // TODO: check NeuVector first to see if it works.

### CEL expressions

The following experiments were conducted in the CEL playground. We should later try them in Kubewarden.

<details><summary>operator = containsAll</summary>

    ```
    //"costcenter" in object.spec.template.metadata.labels && object.spec.template.metadata.labels["costcenter"].matches("^aaa")

    // scenario 1 - only label key is used
    // value = ["badlabel1","badlabel2","badlabel3"]   
    !["badlabel1","badlabel2","badlabel3"].all(x, x in object.spec.template.metadata.labels)

    // scenario 2a - label key and value
    // value = ["badlabel1=badvalue1"]  
    !("badlabel1" in object.spec.template.metadata.labels && 
    object.spec.template.metadata.labels["badlabel1"]=="badvalue1")

    // scenario 2b - label key and value
    // if we have multiple value
    // value = ["badlabel1=badvalue1", "badlabel2=badvalue2"]  
    !(("badlabel1" in object.spec.template.metadata.labels && object.spec.template.metadata.labels["badlabel1"]=="badvalue1")
        &&
    ("badlabel2" in object.spec.template.metadata.labels && object.spec.template.metadata.labels["badlabel2"]=="badvalue2"))

    // scenario 3a - label key and regex value
    // value = ["badlabel1=bad*"]  
    !("badlabel1" in object.spec.template.metadata.labels && 
    object.spec.template.metadata.labels["badlabel1"].matches("^bad.+"))

    // TODO:
    // scenario 3b - multiple label key and regex value example
    // value = ["badlabel1=^bad*", "badlabel2=^anotherbad*"]  

    // TODO:
    // scenario 4  - mixed type of values
    // values = ["badlabel1", "badlabel2=badvalue2"]


    // Some regex notes
    ^bad* matches any string starting with "bad" and optionally followed by "d"s (including the case where "bad" is followed by no "d"s at all, as in "ba").

    ^bad.+ ensures that the string starts with "bad" and is followed by at least one character (not just "bad" itself).
    ```

</details>

<details><summary>operator = containsAny</summary>

```
// scenario 1 - only label key is used
// value = ["badlabel1","badlabel2","badlabel3"]   
!["badlabel1","badlabel2","badlabel3"].exists(x, x in object.spec.template.metadata.labels)

// scenario 2a - label key and value
// value = ["badlabel1=badvalue1"]  
!("badlabel1" in object.spec.template.metadata.labels && object.spec.template.metadata.labels["badlabel1"]=="badvalue1")

// scenario 2b - label key and value
// if we have multiple value
// value = ["badlabel1=badvalue1", "badlabel2=badvalue2"]  
// (same as containsAll, but uses || )
!(("badlabel1" in object.spec.template.metadata.labels && object.spec.template.metadata.labels["badlabel1"]=="badvalue1")
    ||
("badlabel2" in object.spec.template.metadata.labels && object.spec.template.metadata.labels["badlabel2"]=="badvalue2"))
```
</details>

<details><summary>operator = notContainsAny</summary>

```
TODO:
```
</details>

<details><summary>operator = containsOtherThan</summary>

```
// scenario 1 - only label key is used
// value = ["app", "good1","good2","good3"]  
//         only these four keys can be used in the resource, other keys will be denied 
!object.spec.template.metadata.labels.exists(item, !(item in ["app", "good1","good2","good3"]))


```
</details>

## TODO -- need to try, 

```
# operator=containsOtherThan, 
# value=key and value

apiVersion: admissionregistration.k8s.io/v1alpha1
kind: ValidatingAdmissionPolicy
metadata:
  name: validate-annotations
spec:
  paramKind:
    apiVersion: v1
    kind: ConfigMap
  matchConstraints:
    resourceRules:
    - apiGroups: [""]
      apiVersions: ["v1"]
      operations: ["CREATE", "UPDATE"]
      resources: ["pods"]
  validations:
  - expression: 'request.object.metadata.annotations.all(key, key in whitelist && request.object.metadata.annotations[key] == whitelist[key])'
    variables:
    - name: whitelist
      value: '{"key1": "value1", "key2": "value2"}'

```


```
# operator=containsOtherThan, 
# value=regex

apiVersion: admissionregistration.k8s.io/v1alpha1
kind: ValidatingAdmissionPolicy
metadata:
  name: validate-annotations
spec:
  paramKind:
    apiVersion: v1
    kind: ConfigMap
  matchConstraints:
    resourceRules:
    - apiGroups: [""]
      apiVersions: ["v1"]
      operations: ["CREATE", "UPDATE"]
      resources: ["pods"]
  validations:
  - expression: 'request.object.metadata.annotations.all(key, key in whitelist && request.object.metadata.annotations[key].matches(whitelist[key]))'
    variables:
    - name: whitelist
      value: '{"key1": "^[a-zA-Z0-9_-]+$", "key2": "^\\d{4}-\\d{2}-\\d{2}$"}'

```