package pgproject

import "vexentra-api/internal/modules/project"

type projectStatusModel struct {
	Status         string `gorm:"column:status;primaryKey"`
	LabelTH        string `gorm:"column:label_th"`
	Phase          string `gorm:"column:phase"`
	SortOrder      int    `gorm:"column:sort_order"`
	IsTerminal     bool   `gorm:"column:is_terminal"`
	RequiresClient bool   `gorm:"column:requires_client"`
	IsActive       bool   `gorm:"column:is_active"`
}

func (projectStatusModel) TableName() string { return "project_statuses" }

func (m projectStatusModel) ToEntity() project.ProjectStatusMeta {
	return project.ProjectStatusMeta{
		Status:         project.ProjectStatus(m.Status),
		LabelTH:        m.LabelTH,
		Phase:          project.ProjectStatusPhase(m.Phase),
		SortOrder:      m.SortOrder,
		IsTerminal:     m.IsTerminal,
		RequiresClient: m.RequiresClient,
		IsActive:       m.IsActive,
	}
}
