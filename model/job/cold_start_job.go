package job

import (
	model "github.com/AlcEccentric/beck-mizuki/model"
)

type ColdStartOrchJob struct {
	Subjects []model.Subject
	UserIds  []string
}
