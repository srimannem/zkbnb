package block

import (
	"log"

	table "github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/tx"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"gorm.io/gorm"
)

type block struct {
	Txs        []*tx.Tx `gorm:"foreignkey:BlockId"`
	cachedConn sqlc.CachedConn
	table      string
	db         *gorm.DB
	redisConn  *redis.Redis
	cache      multcache.MultCache
}

/*
	Func: GetExecutedBlocksCount
	Params:
	Return: count int64, err error
	Description:  For API /api/v1/info/getLayer2BasicInfo
*/
func (m *block) GetExecutedBlocksCount() (count int64, err error) {
	result, err := m.cache.GetWithSet("cache::block:executed_count", count,
		multcache.SqlQueryCount, m.db, m.table,
		"block_status = ? and deleted_at is NULL", StatusExecuted)
	if err != nil {
		return 0, err
	}
	count, ok := result.(int64)
	if !ok {
		log.Fatal("Error type!")
	}
	return count, nil
}

/*
	Func: GetCommitedBlocksCount
	Params:
	Return: count int64, err error
	Description:  For API /api/v1/info/getLayer2BasicInfo
*/
func (m *block) GetCommitedBlocksCount() (count int64, err error) {
	result, err := m.cache.GetWithSet("cache::block:committed_count", count,
		multcache.SqlQueryCount, m.db, m.table,
		"block_status >= ? and deleted_at is NULL", StatusCommitted)
	if err != nil {
		return 0, err
	}
	count, ok := result.(int64)
	if !ok {
		log.Fatal("Error type!")
	}
	return count, nil
}

/*
	Func: GetBlockByBlockHeight
	Params: blockHeight int64
	Return: err error
	Description:  For API /api/v1/block/getBlockByBlockHeight
*/
func (m *block) GetBlockByBlockHeight(blockHeight int64) (block *table.Block, err error) {
	txForeignKeyColumn := `Txs`
	dbTx := m.db.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	err = m.db.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
	if err != nil {
		return nil, err
	}
	return block, nil
}