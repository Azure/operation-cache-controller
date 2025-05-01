# Cache Controller

The cache controller is responsible for managing the cache resources. It watches the cache resources and updates the cache status based on the cache resource status. The cache controller is responsible for creating, updating, and deleting the cache resources.

## Cache Controller Reconcile Sequence Diagram

::: mermaid
sequenceDiagram
    title: Create Cache Controller Sequence Diagram
    participant user as User
    participant k8s
    participant cc as Cache Controller
    participant oc as Operation Controller
    participant adc as AppDeployment Controller
    participant kcs as Keepalive Count Service

    user ->>+ k8s: Create Cache CR
    k8s -->>- user: Cache CR created successfully
    cc ->>+ k8s: List & Watch Cache CR
    k8s -->> cc: New Cache CR Added
    cc ->>+ cc: Reconciling the Cache
    cc ->>+ k8s: Check if the Cache is ready
    cc ->> kcs: Get the Keepalive Count of current Cache
    kcs -->> cc: Keepalive Count is 3
    cc ->> k8s: list current available cached operations
    k8s -->> cc: List of cached operations
    alt Available Cache Count < Keepalive Count
        cc ->>+ k8s: create (Keepalive Count - Available Cache Count) cached Operation CRs
        k8s -->> cc: Cached Operation CRs created successfully
        oc ->>+ k8s: list & watch the Operation CR
        k8s -->> oc: New Operation CR Added
        oc ->>+ oc: Reconciling the Operation
        oc ->> k8s: Create AppDeployment CRs
        adc ->>+ k8s: list & watch the AppDeployment CR
        k8s -->> adc: New AppDeployment CR Added
        adc ->>+ adc: Reconciling the AppDeployment
        adc ->>+ k8s: Check if all AppDeployments are ready
        k8s -->> adc: All AppDeployments are ready
        adc -->>- adc: AppDeployment reconciled successfully
        oc -->>- oc: Operation reconciled successfully
    else Available Cache Count > Keepalive Count
        cc ->>+ k8s: delete the (Available Cache Count - Keepalive Count) cached Operation CRs
        oc ->>+ k8s: list & watch the Operation CR
        k8s -->> oc: Operation CR Deleted
        oc ->>+ oc: Finalizing the Operation
        oc ->> k8s: cascade delete AppDeployment CRs
        adc ->>+ k8s: list & watch the AppDeployment CR
        k8s -->> adc: AppDeployment CR Deleted
        adc ->>+ adc: Finalizing the AppDeployment
        adc ->> k8s: Clean AppDeployment finalizer annotation
        k8s -->> adc: AppDeployment finalizer annotation cleaned successfully
        adc -->>- adc: AppDeployment finalized successfully
        oc ->> k8s: Clean Operation finalizer annotation
        k8s -->> oc: Operation finalizer annotation cleaned successfully
        oc -->>- oc: Operation finalized successfully
    end
    cc -->>- cc: Cache reconciled successfully
:::

## Cache Controller Finalize Sequence Diagram

::: mermaid
sequenceDiagram
    title: Delete Cache Controller Sequence Diagram
    participant user as User
    participant k8s
    participant cc as Cache Controller
    participant oc as Operation Controller
    participant adc as AppDeployment Controller

    user ->>+ k8s: Delete Cache CR
    cc ->>+ k8s: List & Watch Cache CR
    k8s -->>- cc: New Deleted Cache CR
    cc ->>+ cc: Finalizing the Cache
    cc ->> k8s: list current child Operation CRs
    k8s -->> cc: Return list of cached Operation CRs
    opt if operation was acquired
        cc ->>+ k8s: remove the ownership of the Operation CRs
        k8s -->>- cc: Operation CRs ownership deleted successfully
    end
    cc ->>+ k8s: cascade delete the cached child Operation CRs
    oc ->>+ k8s: list & watch the Operation CR
    k8s -->>- oc: New Operation CR Deleted
    oc ->>+ oc: Finalizing the Operation
    oc ->> k8s: cascade delete AppDeployment CRs
    adc ->>+ k8s: list & watch the AppDeployment CR
    k8s -->> adc: AppDeployment CR Deleted
    adc ->>+ adc: Finalizing the AppDeployment
    adc ->> k8s: Clean AppDeployment finalizer annotation
    k8s -->> adc: AppDeployment finalizer annotation cleaned successfully
    adc -->>- adc: AppDeployment finalized successfully
    oc ->> k8s: Clean Operation finalizer annotation
    k8s -->> oc: Operation finalizer annotation cleaned successfully
    oc -->>- oc: Operation finalized successfully
    cc -->>- cc: Cache finalized successfully
    k8s -->>- cc: Cached Operation CRs deleted successfully

    k8s -->>- user: Cache CR deleted successfully
:::
