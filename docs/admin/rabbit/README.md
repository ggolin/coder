# Tiered RBAC (Enterprise)

> Note: This feature is in active development and behind a closed beta. If you're interested in trying this, contact your account team.

Tiered RBAC is a suite of features that improves Coder's security posture for highly regulated organizations, particularly with multiple business units or security levels.

This includes:

- Organizations
- Groups
- Service Accounts
- Custom Roles

## Organizations

Organizations allow multiple platform teams to coexist within a single Coder instance with isolated cloud credentials, templates, and Coder provisioners. Unlike groups, users within an organization can hold roles that allow them to access all resources within the organization, but not the Coder deployment.

Examples:

- Contractors are isolated to their own `External Developers` organization so they can't access resources in other organizations
- The `ML Platform` and `Developer Platform` teams coexist on a single Coder deployment but bring different provisioners and credentials

Users can belong to multiple organizations. Workspaces, templates, provisioners, and groups are scoped to a single organization.

## Groups

Groups are a logical grouping of users inside an organization. They inherently do not have permissions or roles but, like users, groups can own workspaces and provisioners.

Examples:

- The `@quality-assurance` group has a Windows workspace they use share for testing
- The `@data-science` group has a provisioner in their own Kubernetes cluster but use a standard template

## Service Accounts

Service accounts can be used for CI jobs, third-party integrations, and other automation. Unlike other accounts in Coder, service accounts do not consume a license seat or have an OIDC/password login method, so they cannot be used to log in to the Coder UI.

## Custom Roles

Custom roles can be created to give users a granular set of permissions within the Coder deployment or organization.

Examples:

- The `Banking Compliance Auditor` custom role cannot create workspaces, but can read template source code and view audit logs
- The `Organization Lead` role can access user workspaces for troubleshooting purposes, but cannot edit templates
- The `Platform Member` role cannot edit or create workspaces as they are created via a third-party system

Custom roles can also be applied to service accounts:

- A `Health Check` role can view deployment status but cannot create workspaces, manage templates, or view users
- A `CI` role can update manage templates but cannot create workspaces or view users

---

### Recipes

Learn how to use tiered RBAC with concrete examples:

- [Bring your own cluster](./guides/byoc.md): Allow different groups to bring their own Kubernetes cluster into Coder, or optionally build their own templates on entirely different infrastructure.
- [Shared workspaces](#): Allow users to share workspaces with other users in the same organization.