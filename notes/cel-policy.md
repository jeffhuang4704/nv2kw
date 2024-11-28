## cel policy notes

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

[CEL playground](https://playcel.undistro.io/)

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

### Label criteria

🚧 TODO: need to do multiple values in scenario 2 and scenario 3.

* 🔴 operator = containsAll
    * 1️⃣ scenario 1 - only label key is used, example: value = ["badlabel1","badlabel2","badlabel3"]   
    * 2️⃣ scenario 2 - label key and value, example: value = ["badlabel1=badvalue1"]  
    * 3️⃣ scenario 3 - label key and regex value, example: value = ["badlabel1=^bad*"]  
* 🔴 operator = containsAny
    * 1️⃣ scenario 1 - only label key is used, example: value = ["badlabel1","badlabel2","badlabel3"]   
    * 2️⃣ scenario 2 - label key and value, example: value = ["badlabel1=badvalue1"]  
    * 3️⃣ scenario 3 - label key and regex value, example: value = ["badlabel1=^bad*"]  
* 🔴 operator = notContainsAny
    * 1️⃣ scenario 1 - only label key is used, example: value = ["badlabel1","badlabel2","badlabel3"]   
    * 2️⃣ scenario 2 - label key and value, example: value = ["badlabel1=badvalue1"]  
    * 3️⃣ scenario 3 - label key and regex value, example: value = ["badlabel1=^bad*"]  
* 🔴 operator = containsOtherThan
    * 1️⃣ scenario 1 - only label key is used, example: value = ["badlabel1","badlabel2","badlabel3"]   
    * 2️⃣ scenario 2 - label key and value, example: value = ["badlabel1=badvalue1"]  
    * 3️⃣ scenario 3 - label key and regex value, example: value = ["badlabel1=^bad*"]  

<details><summary>operator = containsAll</summary>

```
//"costcenter" in object.spec.template.metadata.labels && object.spec.template.metadata.labels["costcenter"].matches("^aaa")

// 1️⃣ scenario 1 - only label key is used
// operator = containsAll
// value = ["badlabel1","badlabel2","badlabel3"]   
// need to negate the final value, 
!["badlabel1","badlabel2","badlabel3"].all(x, x in object.spec.template.metadata.labels)

// 2️⃣ scenario 2 - label key and value
// neuvector criteria : labels (env and annotation should be also okay)
// operator = containsAll
// value = ["badlabel1=badvalue1"]  
// need to negate the final value, 
!("badlabel1" in object.spec.template.metadata.labels && 
object.spec.template.metadata.labels["badlabel1"]=="badvalue1")

// if we have multiple value
// value = ["badlabel1=badvalue1", "badlabel2=badvalue2"]  
!(("badlabel1" in object.spec.template.metadata.labels && object.spec.template.metadata.labels["badlabel1"]=="badvalue1")
    &&
("badlabel2" in object.spec.template.metadata.labels && object.spec.template.metadata.labels["badlabel2"]=="badvalue2"))

// 3️⃣ scenario 3 - label key and regex value
// operator = containsAll
// value = ["badlabel1=bad*"]  
//"costcenter" in object.spec.template.metadata.labels && object.spec.template.metadata.labels["costcenter"].matches("^aaa")
!("badlabel1" in object.spec.template.metadata.labels && 
object.spec.template.metadata.labels["badlabel1"].matches("^bad.+"))

// Some regex notes
^bad* matches any string starting with "bad" and optionally followed by "d"s (including the case where "bad" is followed by no "d"s at all, as in "ba").

^bad.+ ensures that the string starts with "bad" and is followed by at least one character (not just "bad" itself).
```

</details>

<details><summary>operator = containsAny</summary>

```
TODO:
```
</details>

<details><summary>operator = notContainsAny</summary>

```
TODO:
```
</details>

<details><summary>operator = containsOtherThan</summary>

```
TODO:
```
</details>