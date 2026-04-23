function cleanRoleList(ctx) {
  const roles = ctx && ctx.member && Array.isArray(ctx.member.roles) ? ctx.member.roles : []
  return roles.map((role) => String(role || "").trim()).filter(Boolean)
}

function hasAnyRole(ctx, roleIds) {
  const roles = cleanRoleList(ctx)
  const allowed = Array.isArray(roleIds) ? roleIds : [roleIds]
  return allowed.map((roleId) => String(roleId || "").trim()).filter(Boolean).some((roleId) => roles.includes(roleId))
}

function canManageShows(ctx) {
  const config = ctx && ctx.config ? ctx.config : {}
  return hasAnyRole(ctx, [config.adminRoleId, config.bookerRoleId])
}

function canAdminOnly(ctx) {
  const config = ctx && ctx.config ? ctx.config : {}
  return hasAnyRole(ctx, config.adminRoleId)
}

function permissionDenied() {
  return { content: "❌ You don't have permission to use this command.", ephemeral: true }
}

module.exports = {
  cleanRoleList,
  hasAnyRole,
  canManageShows,
  canAdminOnly,
  permissionDenied,
}
