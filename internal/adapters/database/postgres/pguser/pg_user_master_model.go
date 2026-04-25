package pguser

import "vexentra-api/internal/modules/user"

type userRoleMasterModel struct {
	Code      string `gorm:"column:code;primaryKey"`
	LabelTH   string `gorm:"column:label_th"`
	LabelEN   string `gorm:"column:label_en"`
	SortOrder int    `gorm:"column:sort_order"`
	IsActive  bool   `gorm:"column:is_active"`
}

func (userRoleMasterModel) TableName() string { return "user_role_master" }

func (m userRoleMasterModel) ToEntity() user.UserRoleMaster {
	return user.UserRoleMaster{
		Code:      m.Code,
		LabelTH:   m.LabelTH,
		LabelEN:   m.LabelEN,
		SortOrder: m.SortOrder,
		IsActive:  m.IsActive,
	}
}

type userStatusMasterModel struct {
	Code      string `gorm:"column:code;primaryKey"`
	LabelTH   string `gorm:"column:label_th"`
	LabelEN   string `gorm:"column:label_en"`
	SortOrder int    `gorm:"column:sort_order"`
	IsActive  bool   `gorm:"column:is_active"`
}

func (userStatusMasterModel) TableName() string { return "user_status_master" }

func (m userStatusMasterModel) ToEntity() user.UserStatusMaster {
	return user.UserStatusMaster{
		Code:      m.Code,
		LabelTH:   m.LabelTH,
		LabelEN:   m.LabelEN,
		SortOrder: m.SortOrder,
		IsActive:  m.IsActive,
	}
}
