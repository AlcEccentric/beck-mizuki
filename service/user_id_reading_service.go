package service

import (
	"math"
	"sync"

	dao "github.com/AlcEccentric/beck-mizuki/dao"
	table "github.com/AlcEccentric/beck-mizuki/model/gen/beck-konomi/public/table"
	job "github.com/AlcEccentric/beck-mizuki/model/job"
	"github.com/rs/zerolog/log"
)

type UserIdReadingService struct {
	konomiAccessor dao.KonomiAccessor
}

func NewUserIdReadingService(accessor dao.KonomiAccessor) *UserIdReadingService {
	return &UserIdReadingService{
		konomiAccessor: accessor,
	}
}

func (svc *UserIdReadingService) GetUserIdReader(numOfUserIdReaders int) func(put func(*job.RegularUpdateOrchJob)) error {
	return func(put func(*job.RegularUpdateOrchJob)) error {
		log.Info().Msg("Reading user ids from database")

		totalUserCnt, err := svc.konomiAccessor.GetRowCount(table.BgmUser)
		if err != nil {
			return err
		}

		userCntPerReader := int(math.Ceil(float64(totalUserCnt) / float64(numOfUserIdReaders)))
		var wg sync.WaitGroup
		for i := 0; i < totalUserCnt; i += userCntPerReader {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				readLimit := userCntPerReader
				if index+userCntPerReader > totalUserCnt {
					readLimit = totalUserCnt - index
				}
				uids, err := svc.konomiAccessor.GetUserIdsPaginated(index, readLimit)

				if err == nil {
					j := &job.RegularUpdateOrchJob{
						UserIds: uids,
					}
					put(j)
				} else {
					log.Error().Err(err).Msgf("Failed to get user ids with offset: %d limit: %d", index, readLimit)
				}
			}(i)
		}
		wg.Wait()
		return nil
	}
}
