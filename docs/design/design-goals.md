# General Design Goals
Whether designing new features or redesigning old ones, AALM should maintain these general design goals:

## Operators are created in other namespace
When an user wants to use an OLM Operator, requires high level priviledges in order to create CRDs or (Cluster)RoleBindings.

In order to manage the access to this priviledges roles without providing them to the final user, all the operators are created in a new namespace that is called `[user-namespace]-operators` where the following resources are created:

- _OperatorGroup_: this resource provides multitenancy configuration. More information can be found [here](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/design/operatorgroups.md). In summary, the `OperatorGroup` define where the OLM operators will watch for their Custom Resources. AALM configures the OperatorGroup to point to the `[user-namespace]`.
- _Subscription_: depending on the `OperatedAsset` that has been created, a new OLM Subscription is created. When this Subscription is created, OLM will creates the operator of the Asset in order to manage all Custom Resources of this specific Asset.

This principle allows the usage of the OLM operators without high permissions but with all advantages of OLM.

## Build (reasonably) small, sharp features
All designs should be kept as simple and tightly scoped as is reasonably possible. Where possible, attempt to follow the unix philosophy for writing small, sharp features that compose well. Additionally, when writing a new proposal, always remember to keep it simple ...student!