function cleanRoleList(ctx) {
  const roles = ctx && ctx.member && Array.isArray(ctx.member.roles) ? ctx.member.roles : []
  return roles.map((role) => String(role || "").trim()).filter(Boolean)
}

function normalizeRoleIds(roleIds) {
  const list = Array.isArray(roleIds) ? roleIds : [roleIds]
  return list.map((roleId) => String(roleId || "").trim()).filter(Boolean)
}

function currentUserId(ctx) {
  return String((ctx && ctx.user && ctx.user.id) || (ctx && ctx.member && ctx.member.id) || "").trim()
}

function currentMemberId(ctx) {
  return String((ctx && ctx.member && ctx.member.id) || "").trim()
}

function hasAnyRole(ctx, roleIds) {
  const roles = cleanRoleList(ctx)
  const allowed = normalizeRoleIds(roleIds)
  return allowed.some((roleId) => roles.includes(roleId))
}

function canManageShows(ctx) {
  const config = ctx && ctx.config ? ctx.config : {}
  return hasAnyRole(ctx, [config.adminRoleId, config.bookerRoleId])
}

function canAdminOnly(ctx) {
  const config = ctx && ctx.config ? ctx.config : {}
  return hasAnyRole(ctx, config.adminRoleId)
}

function permissionDenied(ctx, options) {
  const memberRoles = cleanRoleList(ctx)
  const config = ctx && ctx.config ? ctx.config : {}
  const requiredRoles = normalizeRoleIds((options && options.requiredRoleIds) || [config.adminRoleId, config.bookerRoleId])
  const matchingRoles = requiredRoles.filter((roleId) => memberRoles.includes(roleId))
  const missingRoles = requiredRoles.filter((roleId) => !memberRoles.includes(roleId))
  const requirementLabel = String((options && options.requirementLabel) || "configured allowed role IDs").trim() || "configured allowed role IDs"

  let content = "❌ You don't have permission to use this command."
  content += `\n\nRequirement checked: ${requirementLabel}.`
  content += `\nUser ID: ${currentUserId(ctx) || "(unknown)"}.`
  content += `\nMember ID: ${currentMemberId(ctx) || "(none)"}.`
  content += `\nThe bot sees your member role IDs: ${memberRoles.length > 0 ? memberRoles.join(", ") : "(none)"}.`
  content += `\nRequired role IDs for this command: ${requiredRoles.length > 0 ? requiredRoles.join(", ") : "(not configured)"}.`
  content += `\nExact matching role IDs: ${matchingRoles.length > 0 ? matchingRoles.join(", ") : "(none)"}.`
  content += `\nRequired role IDs not seen on your member object: ${missingRoles.length > 0 ? missingRoles.join(", ") : "(none)"}.`
  content += matchingRoles.length > 0
    ? "\nWhy denied: the bot does see at least one required role ID here, so if this command still failed the next thing to verify is the specific gate for this command and the live runtime config."
    : "\nWhy denied: none of the required role IDs for this command appear on your member object, so the role gate rejected the request."
  content += "\nIf the role names look correct but this still fails, double-check that the configured IDs match the real Discord role IDs."
  content += "\nIf you started the bot with --debug, /debug, /debug-my-roles, and /debug-roles can help you compare what the bot sees against the configured role IDs."

  return { content, ephemeral: true }
}

module.exports = {
  cleanRoleList,
  hasAnyRole,
  canManageShows,
  canAdminOnly,
  permissionDenied,
}
