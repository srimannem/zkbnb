package liquidityoperator

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/monitor/internal/svc"
)

//go:generate mockgen -source api.go -destination api_mock.go -package liquidity

type Model interface {
	CreateLiquidities(pendingNewLiquidityInfos []*liquidity.Liquidity) (err error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: `liquidity`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}