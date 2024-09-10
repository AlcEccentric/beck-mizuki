package service

import (
	dao "github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/model/job"
)

type UserCleaningService struct {
	konomiAccessor dao.KonomiAccessor
}

func NewUserCleaningService(konomiAccessor dao.KonomiAccessor) *UserCleaningService {
	return &UserCleaningService{
		konomiAccessor: konomiAccessor,
	}
}

func (svc *UserCleaningService) GetUserCleaner() func(in *job.RegularUpdateOrchJob) (*job.RegularUpdateOrchJob, error) {
	return func(in *job.RegularUpdateOrchJob) (*job.RegularUpdateOrchJob, error) {
		for _, uid := range in.InactiveUserIds {
			svc.konomiAccessor.DeleteCollectionByUid(uid)
			svc.konomiAccessor.DeleteUser(uid)
		}
		return in, nil
	}
}
