# Operation Controller

## Reconcile Sequence Diagram

``` mermaid
---
title: "Operation Reconcile Sequence Diagram"
---

sequenceDiagram
    participant user as User
    participant k8s
    participant oc as Operation Controller
    participant adc as AppDeployment Controller

    user ->>+ k8s: Create Operation CR
    k8s -->>- user: Operation CR created successfully
    user ->>+ k8s: check if the Operation is ready

    oc ->> k8s: list & watch the Operation CR
    k8s -->> oc: New Operation CR Added
    oc ->>+ oc: reconciling the Operation
    oc ->> oc: generate DeployID for the Operation
    oc ->> k8s: Create all child AppDeployment CR
    k8s -->> oc: All AppDeployment CR created successfully
    adc ->> k8s: list & watch the AppDeployment CR
    k8s -->> adc: New AppDeployment CR Added
    adc ->>+ adc: reconciling the AppDeployment
    adc ->> k8s: update the AppDeployment CR status to ready
    k8s -->> adc: AppDeployment CR status updated successfully
    adc -->>- adc: AppDeployment reconciled successfully

    oc -->>- oc: Operation reconciled successfully
    oc ->> k8s: update the Operation CR status to ready
    k8s -->> oc: Operation CR status updated successfully
    k8s -->>- user: Operation is ready

```

## Finalize Sequence Diagram

``` mermaid

---
title: "Operation Finalize Sequence Diagram"
---

sequenceDiagram
    participant user as User
    participant k8s
    participant oc as Operation Controller
    participant adc as AppDeployment Controller

    alt User deleted
        user ->>+ k8s: Delete Operation CR
        k8s -->>- user: Operation CR deleted successfully
        oc ->> k8s: list & watch the Operation CR
        k8s -->> oc: New Operation CR Deleted
    else Expired
        oc ->> k8s: list & watch the Operation CR
        k8s -->> oc: Operation CR Expired
    end
    oc ->>+ oc: Finalizing the Operation
    oc ->> k8s: cascade delete all child AppDeployment CR
    adc ->> k8s: list & watch the AppDeployment CR
    k8s -->> adc: New AppDeployment CR Deleted
    adc ->>+ adc: Finalizing the AppDeployment
    adc -->> adc: AppDeployment finalized successfully
    adc ->> k8s: Clean AppDeployment CR finalizer annotation
    k8s -->> adc: AppDeployment finalizer annotation cleaned successfully
    adc -->>- adc: AppDeployment deleted successfully
    
    k8s -->> oc: All AppDeployment CR deleted successfully
    oc ->> k8s: Clean Operation finalizer annotation
    k8s -->> oc: Operation finalizer annotation cleaned successfully
    oc -->>- oc: Operation deleted successfully

```
