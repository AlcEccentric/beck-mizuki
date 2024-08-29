package main

import (
	"github.com/alceccentric/beck-crawler/dao"
	"github.com/alceccentric/beck-crawler/orch"
)

func main() {
	bgmClient := dao.NewBgmApiAccessor()
	konomiAccessor := dao.NewKonomiAccessor()
	orch := orch.NewColdStartOrchestrator(bgmClient, konomiAccessor)
	orch.Run(1, 1, 1)
	konomiAccessor.Disconnect()
}
