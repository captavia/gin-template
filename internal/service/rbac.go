package service

import (
	"errors"
	"log"
	"sync"

	"github.com/mikespook/gorbac/v3"
	"github.com/samber/do"
	"gorm.io/gorm"

	"template/internal/model"
)

type RBACService struct {
	db          *gorm.DB
	rwm         *sync.RWMutex
	rbacManager *gorbac.RBAC[uint]
}

func NewRBACService(i *do.Injector) (*RBACService, error) {
	s := &RBACService{
		db:          do.MustInvoke[*gorm.DB](i),
		rwm:         &sync.RWMutex{},
		rbacManager: gorbac.New[uint](),
	}
	s.LoadFromDB()

	return s, nil
}

func (s *RBACService) LoadFromDB() {
	var roles []model.Role
	if err := s.db.
		Preload("Permissions").
		Preload("Parents").
		Find(&roles).Error; err != nil {
		log.Printf("[RBAC] Failed to load roles from DB: %v", err)
		return
	}

	newManager := gorbac.New[uint]()

	// 第一步：注册所有角色及其直接权限
	for _, r := range roles {
		role := gorbac.NewRole(r.ID)
		for _, p := range r.Permissions {
			role.Assign(gorbac.NewPermission(p.ID))
		}
		newManager.Add(role)
	}

	// 第二步：设置父角色继承关系
	for _, r := range roles {
		for _, parent := range r.Parents {
			if err := newManager.SetParent(r.ID, parent.ID); err != nil {
				log.Printf("[RBAC] Failed to set parent %d for role %d: %v", parent.ID, r.ID, err)
			}
		}
	}

	s.rwm.Lock()
	s.rbacManager = newManager
	s.rwm.Unlock()

	log.Printf("[RBAC] Loaded %d roles from DB", len(roles))
}

// IsGranted 检查指定角色是否拥有某个权限（含继承）。
func (s *RBACService) IsGranted(roleID uint, permissionID uint) bool {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	return s.rbacManager.IsGranted(roleID, gorbac.NewPermission(permissionID), nil)
}

// Reload 重新从数据库加载 RBAC 数据，可在权限变更后调用。
func (s *RBACService) Reload() {
	s.LoadFromDB()
}

// ── 角色管理 ─────────────────────────────────────────────────────────────────

// CreateRole 创建一个新角色并持久化到数据库。
func (s *RBACService) CreateRole(code, name string) (*model.Role, error) {
	role := &model.Role{Code: code, Name: name}
	if err := s.db.Create(role).Error; err != nil {
		return nil, err
	}
	// 同步到内存：Add 内部有锁，外部只需读锁保护指针访问
	s.rwm.RLock()
	s.rbacManager.Add(gorbac.NewRole(role.ID))
	s.rwm.RUnlock()
	return role, nil
}

// ListRoles 返回所有角色列表（含关联权限）。
func (s *RBACService) ListRoles() ([]model.Role, error) {
	var roles []model.Role
	if err := s.db.Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// ── 权限管理 ─────────────────────────────────────────────────────────────────

// CreatePermission 创建一个新权限并持久化到数据库。
func (s *RBACService) CreatePermission(code, name string) (*model.Permission, error) {
	perm := &model.Permission{Code: code, Name: name}
	if err := s.db.Create(perm).Error; err != nil {
		return nil, err
	}
	return perm, nil
}

// ListPermissions 返回所有权限列表。
func (s *RBACService) ListPermissions() ([]model.Permission, error) {
	var perms []model.Permission
	if err := s.db.Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// ── 角色-权限关联管理 ──────────────────────────────────────────────────────────

// AssignPermission 为角色添加权限（数据库 + 内存同步）。
func (s *RBACService) AssignPermission(roleID, permissionID uint) error {
	role, perm, err := s.fetchRoleAndPermission(roleID, permissionID)
	if err != nil {
		return err
	}

	if err := s.db.Model(role).Association("Permissions").Append(perm); err != nil {
		return err
	}

	s.rwm.RLock()
	defer s.rwm.RUnlock()
	if r, _, err := s.rbacManager.Get(roleID); err == nil {
		r.Assign(gorbac.NewPermission(permissionID))
	}
	return nil
}

// RevokePermission 为角色移除权限（数据库 + 内存同步）。
func (s *RBACService) RevokePermission(roleID, permissionID uint) error {
	role, perm, err := s.fetchRoleAndPermission(roleID, permissionID)
	if err != nil {
		return err
	}

	if err := s.db.Model(role).Association("Permissions").Delete(perm); err != nil {
		return err
	}

	s.rwm.RLock()
	defer s.rwm.RUnlock()
	if r, _, err := s.rbacManager.Get(roleID); err == nil {
		r.Revoke(gorbac.NewPermission(permissionID))
	}
	return nil
}

func (s *RBACService) fetchRoleAndPermission(roleID, permissionID uint) (*model.Role, *model.Permission, error) {
	var role model.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("role not found")
		}
		return nil, nil, err
	}

	var perm model.Permission
	if err := s.db.First(&perm, permissionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("permission not found")
		}
		return nil, nil, err
	}

	return &role, &perm, nil
}
