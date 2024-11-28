## cel policy notes

[CEL playground](https://playcel.undistro.io/)

We can leverage the Kubewarden CEL policy to implement the following NeuVector criteria.
- labels
- annotation
- environment variables

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