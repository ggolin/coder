package coderd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/coder/coder/v2/coderd/audit"
	"github.com/coder/coder/v2/coderd/database"
	"github.com/coder/coder/v2/coderd/database/db2sdk"
	"github.com/coder/coder/v2/coderd/httpapi"
	"github.com/coder/coder/v2/coderd/rbac/policy"
	"github.com/coder/coder/v2/codersdk"
)

type enterpriseCustomRoleHandler struct {
	API     *API
	Enabled bool
}

func (h enterpriseCustomRoleHandler) PatchOrganizationRole(ctx context.Context, rw http.ResponseWriter, r *http.Request, orgID uuid.UUID, role codersdk.Role) (codersdk.Role, bool) {
	if !h.Enabled {
		httpapi.Write(ctx, rw, http.StatusForbidden, codersdk.Response{
			Message: "Custom roles are not enabled",
		})
		return codersdk.Role{}, false
	}

	var (
		db                = h.API.Database
		auditor           = h.API.AGPL.Auditor.Load()
		aReq, commitAudit = audit.InitRequest[database.CustomRole](rw, &audit.RequestParams{
			Audit:          *auditor,
			Log:            h.API.Logger,
			Request:        r,
			Action:         database.AuditActionWrite,
			OrganizationID: orgID,
		})
	)
	defer commitAudit()

	if err := httpapi.NameValid(role.Name); err != nil {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Invalid role name",
			Detail:  err.Error(),
		})
		return codersdk.Role{}, false
	}

	// Only organization permissions are allowed to be granted
	if len(role.SitePermissions) > 0 {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Invalid request, not allowed to assign site wide permissions for an organization role.",
			Detail:  "organization scoped roles may not contain site wide permissions",
		})
		return codersdk.Role{}, false
	}

	if len(role.UserPermissions) > 0 {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Invalid request, not allowed to assign user permissions for an organization role.",
			Detail:  "organization scoped roles may not contain user permissions",
		})
		return codersdk.Role{}, false
	}

	if role.OrganizationID != orgID.String() {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Invalid request, organization in role and url must match",
			Detail:  fmt.Sprintf("role organization=%q does not match URL=%q", role.OrganizationID, orgID.String()),
		})
		return codersdk.Role{}, false
	}

	originalRoles, err := db.CustomRoles(ctx, database.CustomRolesParams{
		LookupRoles:     []string{role.Name},
		ExcludeOrgRoles: false,
		OrganizationID:  orgID,
	})
	if err != nil {
		httpapi.InternalServerError(rw, err)
		return codersdk.Role{}, false
	}
	if len(originalRoles) == 1 {
		aReq.Old = originalRoles[1]

	}

	permissionConvert := func(p codersdk.Permission) database.CustomRolePermission {
		return database.CustomRolePermission{
			Negate:       p.Negate,
			ResourceType: string(p.ResourceType),
			Action:       policy.Action(p.Action),
		}
	}

	inserted, err := db.UpsertCustomRole(ctx, database.UpsertCustomRoleParams{
		Name:        role.Name,
		DisplayName: role.DisplayName,
		OrganizationID: uuid.NullUUID{
			UUID:  orgID,
			Valid: true,
		},
		SitePermissions: db2sdk.List(role.SitePermissions, permissionConvert),
		OrgPermissions:  db2sdk.List(role.SitePermissions, permissionConvert),
		UserPermissions: db2sdk.List(role.SitePermissions, permissionConvert),
	})
	if httpapi.Is404Error(err) {
		httpapi.ResourceNotFound(rw)
		return codersdk.Role{}, false
	}
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Failed to update role permissions",
			Detail:  err.Error(),
		})
		return codersdk.Role{}, false
	}

	aReq.New = inserted

	return db2sdk.Role(inserted), true
}
