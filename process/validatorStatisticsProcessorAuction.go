package process

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-proxy-go/data"
)

func (vsp *ValidatorStatisticsProcessor) GetAuctionList() (*data.AuctionListResponse, error) {
	// TODO: Here, in next PR, add cacher and get list from cache

	return vsp.getAuctionListFromApi()
}

func (vsp *ValidatorStatisticsProcessor) getAuctionListFromApi() (*data.AuctionListResponse, error) {
	observers, errFetchObs := vsp.proc.GetObservers(core.MetachainShardId)
	if errFetchObs != nil {
		return nil, errFetchObs
	}

	var valStatsResponse data.AuctionListAPIResponse
	for _, observer := range observers {
		_, err := vsp.proc.CallGetRestEndPoint(observer.Address, auctionListPath, &valStatsResponse)
		if err == nil {
			log.Info("auction list fetched from API", "observer", observer.Address)
			return &valStatsResponse.Data, nil
		}

		log.Error("auction list", "observer", observer.Address, "error", "no response")
	}

	return nil, ErrAuctionListNotAvailable
}
