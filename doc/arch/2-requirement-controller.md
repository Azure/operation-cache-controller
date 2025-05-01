# Requirement Controller

## Cache Hit

``` mermaid
---
title: "Operation Cache Controller Manager Sequence Diagram"
---

sequenceDiagram
    autonumber
    actor user
    participant ctl as deployctl
    participant k8s

    user ->>+ ctl: deployctl create -f <path-to-yaml-file>
    ctl ->>+ k8s: Check if cache CR for exist
    k8s -->>- ctl: Cache CR exist
    ctl ->>+ k8s: Check if available cache items exist
    k8s -->>- ctl: Available cache items exist
    ctl ->>+ k8s: select one cached deployment get deployid and set annotation to acquired
    k8s -->>- ctl: Annotation for cached deployment is set successfully
    ctl -->>- user: return DeployID
```

## Cache Miss

``` mermaid

---
title: "Operation Cache Controller Manager Sequence Diagram"
---

sequenceDiagram
    autonumber
    actor user
    participant ctl as deployctl
    participant k8s

    user ->>+ ctl: deployctl create -f <path-to-yaml-file>
    ctl ->>+ k8s: Check if cache CR for exist
    k8s -->>- ctl: Cache CR does not exist
    ctl ->>+ k8s: Create Cache CR
    k8s -->>- ctl: Cache CR created successfully
    ctl ->>+ k8s: create Deployment CR
    k8s -->>- ctl: Deployment CR created successfully
    ctl ->>+ k8s: check if the Deployment is ready
    k8s -->>- ctl: Deployment is ready
    ctl -->>- user: return DeployID

```
