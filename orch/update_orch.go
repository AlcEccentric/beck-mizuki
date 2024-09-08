package orch

import (
	dao "github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/service"
)

type UpdateOrchestrator struct {
	bgmClient          *dao.BgmApiAccessor
	subjectSvc         *service.SubjectService
	userIdSvc          *service.UserIdService
	persistenceService *service.UserPersistenceService
}

func NewUpdateOrchestrator(bgmClient *dao.BgmApiAccessor, konomiAccessor *dao.KonomiAccessor) *UpdateOrchestrator {
	return &UpdateOrchestrator{
		bgmClient:          bgmClient,
		subjectSvc:         service.NewSubjectService(bgmClient),
		userIdSvc:          service.NewUserIdService(),
		persistenceService: service.NewUserPersistenceService(bgmClient, konomiAccessor),
	}
}

func (orch *UpdateOrchestrator) Run(numOfCollectionUpdater, numOfDataCleaner int) {
	// 1. fetch user id with last active time & divide into batches

	// For user info each batch:
	// 1. perform VIP activity check
	// 2. put user into a inactive user list if the check fails
	// 3. otherwise
	// 3.1. fetch collections happened after last active time for the user
	// 3.2. update last active time
	// 3.3. persist user info & new collections
	// 4. remove user & their collections info for users in inactive user list

	// TODO:
	// As updater will update user info frequently, it ensures user left in the table is VIP
	// Thus, for user alreadly in the table, the cold start orchestrator should not perform VIP testing and should just fetch collections since last active time
}
