package price

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/cache"
)

type Fetcher interface {
	GetCurrencyPrice(ctx context.Context, l2Symbol string) (price float64, err error)
}

func NewFetcher(memCache *cache.MemCache, cmcUrl, cmcToken string) Fetcher {
	return &fetcher{
		memCache: memCache,
		cmcUrl:   cmcUrl,
		cmcToken: cmcToken,
	}
}

type fetcher struct {
	memCache *cache.MemCache
	cmcUrl   string
	cmcToken string
}

func (f *fetcher) GetCurrencyPrice(_ context.Context, symbol string) (float64, error) {
	return f.memCache.GetPriceWithFallback(symbol, func() (interface{}, error) {
		quoteMap, err := f.getLatestQuotes(symbol)
		if err != nil {
			return 0, err
		}
		q, ok := quoteMap[symbol]
		if !ok {
			return 0, nil
		}
		price := q.Quote["USD"].Price
		return price, err
	})
}

func (f *fetcher) getLatestQuotes(symbol string) (map[string]QuoteLatest, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s%s", f.cmcUrl, symbol)
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errorcode.HttpErrFailToRequest
	}
	reqest.Header.Add("X-CMC_PRO_API_KEY", f.cmcToken)
	reqest.Header.Add("Accept", "application/json")
	resp, err := client.Do(reqest)
	if err != nil {
		return nil, errorcode.HttpErrClientDo
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errorcode.IoErrFailToRead
	}
	currencyPrice := &currencyPrice{}
	if err = json.Unmarshal(body, &currencyPrice); err != nil {
		return nil, errorcode.JsonErrUnmarshal
	}
	ifcs, ok := currencyPrice.Data.(interface{})
	if !ok {
		return nil, errors.New("type conversion error")
	}
	quotesLatest := make(map[string]QuoteLatest, 0)
	for _, coinObj := range ifcs.(map[string]interface{}) {
		b, err := json.Marshal(coinObj)
		if err != nil {
			return nil, errorcode.JsonErrMarshal
		}
		quoteLatest := &QuoteLatest{}
		err = json.Unmarshal(b, quoteLatest)
		if err != nil {
			return nil, errorcode.JsonErrUnmarshal
		}
		quotesLatest[quoteLatest.Symbol] = *quoteLatest
	}
	return quotesLatest, nil
}
