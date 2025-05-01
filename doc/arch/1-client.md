# deployctl

deployctl is a command-line tool that interacts with the Operation Cache Controller Manager to create and manage resources in a Kubernetes cluster.

## Create

```shell
deployctl create -f <path-to-yaml-file>
```

## Delete

```shell
deployctl delete -o <operation-id>
```

## Update

```shell
deployctl update -o <operation-id> -f <path-to-yaml-file>
```

## Sequence Diagram interact with Operation Cache Controller Manager

### Without cache

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
    ctl ->>+ k8s: Create Requirement CR
    k8s -->>- ctl: Requirement CR created successfully
    ctl ->>+ k8s: check if the Requirement is ready
    k8s -->>- ctl: Requirement is ready
    ctl -->>- user: return DeployID
```
