## cel policy notes

We can apply the CEL policy to meet the following NeuVector criteria.
- labels
- annotation
- environment variables

The following operators are used in these criteria.

```
    "containsAll"
    "containsAny"
    "notContainsAny"
    "containsOtherThan"
```

The following are some experiments attempting to implement these criteria and their operators. 
For each, we need to consider the following scenarios, using labels as an example.

scenario 1 - only label key is used
scenario 2 - label key and value
scenario 3 - label key and value, and the value has regex

### Label criteria

```
//"costcenter" in object.spec.template.metadata.labels && object.spec.template.metadata.labels["costcenter"].matches("^aaa")

// 1️⃣ scenario 1 - only label key is used
// neuvector criteria : labels (env and annotation should be also okay)
// operator = containsAll
// value = ["badlabel1","badlabel2","badlabel3"]   
// need to negate the final value, 
// in Kubewarden CEL policy => The policy will be evaluated as allowed if all the CEL expressions are evaluated as true
// check if all the values user provided are used in the resource labels..
!["badlabel1","badlabel2","badlabel3"].all(x, x in object.spec.template.metadata.labels)

// 2️⃣ scenario 2 - label key and value
// neuvector criteria : labels (env and annotation should be also okay)
// operator = containsAll
// value = ["badlabel1=badvalue1"]  
// need to negate the final value, 
!("badlabel1" in object.spec.template.metadata.labels && 
object.spec.template.metadata.labels["badlabel1"]=="badvalue1")
```