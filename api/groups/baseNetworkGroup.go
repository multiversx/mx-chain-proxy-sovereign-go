package groups

import (
	"net/http"

	"github.com/ElrondNetwork/elrond-proxy-go/api/shared"
	"github.com/ElrondNetwork/elrond-proxy-go/data"
	"github.com/ElrondNetwork/elrond-proxy-go/process"
	"github.com/gin-gonic/gin"
)

type networkGroup struct {
	facade NetworkFacadeHandler
	*baseGroup
}

// NewNetworkGroup returns a new instance of networkGroup
func NewNetworkGroup(facadeHandler data.FacadeHandler) (*networkGroup, error) {
	facade, ok := facadeHandler.(NetworkFacadeHandler)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	ng := &networkGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	baseRoutesHandlers := []*data.EndpointHandlerData{
		{Path: "/status/:shard", Handler: ng.getNetworkStatusData, Method: http.MethodGet},
		{Path: "/config", Handler: ng.getNetworkConfigData, Method: http.MethodGet},
		{Path: "/economics", Handler: ng.getEconomicsData, Method: http.MethodGet},
		{Path: "/esdts", Handler: ng.getEsdts, Method: http.MethodGet},
		{Path: "/enable-epochs", Handler: ng.getEnableEpochs, Method: http.MethodGet},
	}
	ng.baseGroup.endpoints = baseRoutesHandlers

	return ng, nil
}

// getNetworkStatusData will expose the node network metrics for the given shard
func (group *networkGroup) getNetworkStatusData(c *gin.Context) {
	shardIDUint, err := shared.FetchShardIDFromRequest(c)
	if err != nil {
		shared.RespondWith(c, http.StatusBadRequest, nil, process.ErrInvalidShardId.Error(), data.ReturnCodeRequestError)
		return
	}

	networkStatusResults, err := group.facade.GetNetworkStatusMetrics(shardIDUint)
	if err != nil {
		shared.RespondWith(c, http.StatusInternalServerError, nil, err.Error(), data.ReturnCodeInternalError)
		return
	}

	c.JSON(http.StatusOK, networkStatusResults)
}

// getNetworkConfigData will expose the node network metrics for the given shard
func (group *networkGroup) getNetworkConfigData(c *gin.Context) {
	networkConfigResults, err := group.facade.GetNetworkConfigMetrics()
	if err != nil {
		shared.RespondWith(c, http.StatusInternalServerError, nil, err.Error(), data.ReturnCodeInternalError)
		return
	}

	c.JSON(http.StatusOK, networkConfigResults)
}

// getEconomicsData will expose the economics data metrics from an observer (if any available) in json format
func (group *networkGroup) getEconomicsData(c *gin.Context) {
	economicsData, err := group.facade.GetEconomicsDataMetrics()
	if err != nil {
		shared.RespondWith(c, http.StatusInternalServerError, nil, err.Error(), data.ReturnCodeInternalError)
		return
	}

	c.JSON(http.StatusOK, economicsData)
}

// getEsdts will expose all the issued ESDTs
func (group *networkGroup) getEsdts(c *gin.Context) {
	allIssuedESDTs, err := group.facade.GetAllIssuedESDTs()
	if err != nil {
		shared.RespondWith(c, http.StatusInternalServerError, nil, err.Error(), data.ReturnCodeInternalError)
		return
	}

	c.JSON(http.StatusOK, allIssuedESDTs)
}

func (group *networkGroup) getEnableEpochs(c *gin.Context) {
	enableEpochsMetrics, err := group.facade.GetEnableEpochsMetrics()
	if err != nil {
		shared.RespondWith(c, http.StatusInternalServerError, nil, err.Error(), data.ReturnCodeInternalError)
		return
	}

	c.JSON(http.StatusOK, enableEpochsMetrics)
}
