package job

import (
	model "github.com/alceccentric/beck-crawler/model"
)

type ColdStartOrchJob struct {
	Subjects []model.Subject
	UserIds  []string
}
