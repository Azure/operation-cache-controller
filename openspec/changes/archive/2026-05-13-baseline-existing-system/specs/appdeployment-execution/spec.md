## ADDED Requirements

### Requirement: AppDeployment runs the provision Job on creation

The controller SHALL, when an `AppDeployment` is created, launch a Kubernetes `Job` derived from `spec.provision` owned by the `AppDeployment`, transition `status.phase` from `""` → `Pending` → `Deploying`, and set `status.phase=Ready` when the provision `Job` reports `Complete=True`.

#### Scenario: Successful provision Job promotes AppDeployment to Ready

- **WHEN** an `AppDeployment` is created and its provision `Job` succeeds
- **THEN** `status.phase` of the `AppDeployment` becomes `Ready`

#### Scenario: Failed provision Job is retried and keeps AppDeployment out of Ready

- **WHEN** the provision `Job` reports `Failed=True`
- **THEN** the controller deletes the failed provision `Job`, creates a replacement provision `Job`, and keeps the `AppDeployment` out of `Ready` until a provision `Job` succeeds

### Requirement: AppDeployment respects declared dependencies

The controller SHALL NOT launch the provision `Job` for an `AppDeployment` whose `spec.dependencies` include sibling app names until every dependency `AppDeployment` (sharing the same parent `Operation` via `spec.opId`) has reached `status.phase=Ready`.

#### Scenario: Dependent app waits for its dependency

- **WHEN** `AppDeployment` `app-b` declares `spec.dependencies=["app-a"]` and `app-a` is not yet `Ready`
- **THEN** `app-b` remains in `status.phase=Pending` and no provision `Job` is created for it
- **AND** once `app-a` reaches `status.phase=Ready`, the controller launches `app-b`'s provision `Job`

### Requirement: AppDeployment runs the teardown Job on deletion

The controller SHALL add the finalizer `finalizer.appdeployment.devinfra.goms.io` to every `AppDeployment`, and on deletion SHALL launch a `Job` derived from `spec.teardown` owned by the `AppDeployment`, transition `status.phase` to `Deleting`, and remove the finalizer after `status.phase=Deleted`. The controller sets `status.phase=Deleted` when the teardown `Job` succeeds, or after it observes and deletes a failed teardown `Job` while emitting a warning event.

#### Scenario: Teardown Job runs before AppDeployment is removed

- **WHEN** an `AppDeployment` in `status.phase=Ready` is deleted
- **THEN** a teardown `Job` is created, the `AppDeployment` reports `status.phase=Deleting` until the teardown attempt completes or fails, then reports `status.phase=Deleted` and has its finalizer removed

### Requirement: AppDeployment owns its Jobs for cascade deletion

The controller SHALL set the `AppDeployment` as the controller `ownerReference` of every provision and teardown `Job` it creates, so deleting the `AppDeployment` causes Kubernetes garbage collection of orphaned `Job`s.

#### Scenario: Deleting AppDeployment removes its Jobs

- **WHEN** an `AppDeployment` and its provision `Job` exist, and the `AppDeployment` is deleted
- **THEN** the provision `Job` is garbage-collected by Kubernetes via `ownerReferences`
