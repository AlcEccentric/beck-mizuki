package job

import (
	bgmModel "github.com/alceccentric/beck-crawler/dao/bgm/model"
)

type ColdStartOrchJob struct {
	Subjects []bgmModel.Subject
	UserIds  []string
}
