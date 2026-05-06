package service

import (
	"github.com/casbin/casbin/v3"
	"github.com/samber/do"
)

type PermissionService struct {
	enforcer *casbin.Enforcer
}

func NewPermissionService(i *do.Injector) (*PermissionService, error) {
	return &PermissionService{
		enforcer: do.MustInvoke[*casbin.Enforcer](i),
	}, nil
}

// AddRoleForUser 为用户分配角色
func (s *PermissionService) AddRoleForUser(user string, role string) (bool, error) {
	return s.enforcer.AddGroupingPolicy(user, role)
}

// RemoveRoleForUser 移除用户的角色
func (s *PermissionService) RemoveRoleForUser(user string, role string) (bool, error) {
	return s.enforcer.RemoveGroupingPolicy(user, role)
}

// AddPermissionForRole 为角色添加权限
func (s *PermissionService) AddPermissionForRole(role string, path string, method string) (bool, error) {
	return s.enforcer.AddPolicy(role, path, method)
}

// RemovePermissionForRole 移除角色的权限
func (s *PermissionService) RemovePermissionForRole(role string, path string, method string) (bool, error) {
	return s.enforcer.RemovePolicy(role, path, method)
}

// GetRolesForUser 获取用户拥有的角色
func (s *PermissionService) GetRolesForUser(user string) ([]string, error) {
	return s.enforcer.GetRolesForUser(user)
}

// GetPermissionsForRole 获取角色的权限列表
func (s *PermissionService) GetPermissionsForRole(role string) ([][]string, error) {
	return s.enforcer.GetPermissionsForUser(role)
}
