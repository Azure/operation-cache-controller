# App Deployment

## Provision Sequence Diagram

``` mermaid

sequenceDiagram
    title: App Deployment Provision Sequence Diagram
    participant user as User
    participant k8s
    participant adc as AppDeployment Controller
    participant job as K8S Job Controller
    participant reg as Image Registry

    user ->>+ k8s: Create AppDeployment CR
    k8s -->>- user: AppDeployment CR created successfully
    user ->+ k8s: check if the AppDeployment is ready

    adc ->> k8s: list & watch the AppDeployment CR
    k8s -->> adc: New AppDeployment CR Added

    adc ->>+ adc: reconciling the AppDeployment

    adc ->> k8s: list all dependencies of the AppDeployment
    k8s -->> adc: dependencies listed successfully

    adc ->> adc: check if all dependencies are ready
    adc ->> k8s: create Provision Job CR
    k8s -->> adc: Create Job CR successfully
    adc ->> k8s: check if the Job CR is ready
    job ->> k8s: list & watch the Job CR
    k8s -->> job: New Job CR Added
    job ->>+ job: reconciling the job
    job ->>+ reg: pull the provisioner image
    reg -->>- job: provisioner image pulled successfully
    job ->> job: running the provisioning job
    job ->> k8s: update the Job CR status to completed
    k8s -->> job: Job CR status updated successfully
    job -->>- job: Job reconciled successfully
    k8s -->> adc: Job CR is ready
    adc ->> k8s: update the AppDeployment CR status to ready
    adc -->>- adc: AppDeployment reconciled successfully
    k8s -->>- user: AppDeployment is ready

```

## Teardown Sequence Diagram

``` mermaid

sequenceDiagram
    title: "App Deployment Teardown Sequence Diagram"
    participant user as User
    participant k8s
    participant adc as AppDeployment Controller
    participant job as K8S Job Controller
    participant reg as Image Registry

    user ->>+ k8s: Delete AppDeployment CR
    k8s -->>- user: AppDeployment CR deleted successfully
    user ->+ k8s: check if the AppDeployment is deleted

    adc ->> k8s: list & watch the AppDeployment CR
    k8s -->> adc: AppDeployment CR deleted

    adc ->>+ adc: reconciling the AppDeployment

    adc ->> k8s: create Teardown Job CR
    k8s -->> adc: Create Job CR successfully
    adc ->> k8s: check if the Job CR is ready
    job ->> k8s: list & watch the Job CR
    k8s -->> job: New Job CR Added
    job ->>+ job: reconciling the job
    job ->>+ reg: pull the provisioner image
    reg -->>- job: provisioner image pulled successfully
    job ->> job: running the teardown job
    job ->> k8s: update the Job CR status to completed
    k8s -->> job: Job CR status updated successfully
    job -->>- job: Job reconciled successfully
    k8s -->> adc: Job CR is ready
    adc ->> k8s: update the AppDeployment CR status to deleted
    adc -->>- adc: AppDeployment reconciled successfully
    k8s -->>- user: AppDeployment is deleted
```

## Spec for Job CR

### Provsion Job

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: provision-envid-deployment-1
  namespace: <namespace>
spec:
  template:
    spec:
      containers:
        - name: <container-name>
          image: <image>
          env:
            - name: <env-name>
              value: <env-value>
          command: "create"
      restartPolicy: Never
    backoffLimit: 4
```

### Teardown Job

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: teardown-envid-deployment-1
  namespace: <namespace>
spec:
    template:
        spec:
        containers:
            - name: <container-name>
            image: <image>
            env:
                - name: <env-name>
                value: <env-value>
            command: "delete"
        restartPolicy: Never
        backoffLimit: 4
```
