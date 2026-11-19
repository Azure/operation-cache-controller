## ADDED Requirements

### Requirement: Operation only reconciles AppDeployments it controls

The controller SHALL, when reconciling an `Operation`, consider only `AppDeployment` resources in the `Operation`'s namespace whose controller owner reference matches that `Operation`'s UID. The controller SHALL NOT create, update or delete any other `AppDeployment` as part of that `Operation`'s reconciliation.

#### Scenario: Same-named Operations in different namespaces are independent

- **WHEN** two `Operation`s with the same name exist in namespaces `ns-a` and `ns-b`
- **AND** the `Operation` in `ns-a` owns `AppDeployment`s
- **THEN** reconciling the `Operation` in `ns-b` does not list, update or delete the `AppDeployment`s in `ns-a`

#### Scenario: AppDeployment not controlled by the Operation is ignored

- **WHEN** an `AppDeployment` in the `Operation`'s namespace has a controller owner reference with a different UID
- **THEN** reconciling the `Operation` leaves that `AppDeployment` unchanged
