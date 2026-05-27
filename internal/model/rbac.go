package model

import "gorm.io/gorm"

type Permission struct {
	gorm.Model
	Code string `gorm:"type:varchar(32);uniqueIndex;not null" comment:"权限标识"`
	Name string `gorm:"type:varchar(32)" comment:"权限名称,如 读取用户列表"`
}

// Role 角色表
type Role struct {
	gorm.Model
	Code string `gorm:"type:varchar(32);uniqueIndex;not null" comment:"角色标识,如 admin"`
	Name string `gorm:"type:varchar(32)" comment:"角色名称,如 超级管理员"`

	Permissions []Permission `gorm:"many2many:role_permissions;"`

	Parents []Role `gorm:"many2many:role_parents;joinForeignKey:role_id;joinReferences:parent_id"`
}
