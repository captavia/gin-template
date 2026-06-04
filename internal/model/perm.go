package model

// ── Permissions ──────────────────────────────────────────────────
const (
	PermCreateUser uint = iota + 1
	PermEditUser
	PermDeleteUser
	PermManageRoles
)

// ── Roles ───────────────────────────────────────────────────────
const (
	RoleSuperAdmin uint = iota + 1
	RoleManager
	RoleNormalUser
)
