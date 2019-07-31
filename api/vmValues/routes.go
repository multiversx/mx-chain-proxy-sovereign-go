package vmValues

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"

	apiErrors "github.com/ElrondNetwork/elrond-proxy-go/api/errors"
	"github.com/gin-gonic/gin"
)

// VmValueRequest represents the structure on which user input for generating a new transaction will validate against
type VmValueRequest struct {
	ScAddress string   `form:"scAddress" json:"scAddress"`
	FuncName  string   `form:"funcName" json:"funcName"`
	Args      []string `form:"args"  json:"args"`
}

// Routes defines address related routes
func Routes(router *gin.RouterGroup) {
	router.POST("/hex", GetVmValueAsHexBytes)
	router.POST("/string", GetVmValueAsString)
	router.POST("/int", GetVmValueAsBigInt)
}

func vmValueFromAccount(c *gin.Context, resType string) ([]byte, int, error) {
	ef, ok := c.MustGet("elrondProxyFacade").(FacadeHandler)
	if !ok {
		return nil, http.StatusInternalServerError, apiErrors.ErrInvalidAppContext
	}

	var gval = VmValueRequest{}
	err := c.ShouldBindJSON(&gval)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	argsBuff := make([][]byte, 0)
	for _, arg := range gval.Args {
		buff, err := hex.DecodeString(arg)
		if err != nil {
			return nil,
				http.StatusBadRequest,
				errors.New(fmt.Sprintf("'%s' is not a valid hex string: %s", arg, err.Error()))
		}

		argsBuff = append(argsBuff, buff)
	}

	returnedData, err := ef.GetVmValue(resType, gval.ScAddress, gval.FuncName, argsBuff...)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return returnedData, http.StatusOK, nil
}

// GetVmValueAsHexBytes returns the data as byte slice
func GetVmValueAsHexBytes(c *gin.Context) {
	data, status, err := vmValueFromAccount(c, "hex")
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("get vm value as hex bytes: %s", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": hex.EncodeToString(data)})
}

// GetVmValueAsString returns the data as string
func GetVmValueAsString(c *gin.Context) {
	data, status, err := vmValueFromAccount(c, "string")
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("get vm value as string: %s", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": string(data)})
}

// GetVmValueAsBigInt returns the data as big int
func GetVmValueAsBigInt(c *gin.Context) {
	data, status, err := vmValueFromAccount(c, "int")
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("get vm value as big int: %s", err)})
		return
	}

	value, ok := big.NewInt(0).SetString(string(data), 10)
	if !ok {
		c.JSON(status, gin.H{"error": fmt.Sprintf("value %s could not be converted to a big int", string(data))})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": value.String()})
}
