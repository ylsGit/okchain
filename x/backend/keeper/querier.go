package keeper

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/okex/okchain/x/backend/types"
	"github.com/okex/okchain/x/common"
	orderTypes "github.com/okex/okchain/x/order/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		if keeper.Config.EnableBackend == false {
			response := common.GetErrorResponse(-1, "", "Backend Plugin's Not Enabled")
			res, eJson := json.Marshal(response)
			if eJson != nil {
				return nil, sdk.ErrInternal(eJson.Error())
			}
			return res, nil
		}

		defer func() {
			if e := recover(); e != nil {
				errMsg := fmt.Sprintf("%+v", e)
				response := common.GetErrorResponse(-1, "", errMsg)
				resJson, eJson := json.Marshal(response)
				if eJson != nil {
					res = nil
					err = sdk.ErrInternal(eJson.Error())
				} else {
					res = resJson
					err = nil
				}

			}
		}()

		switch path[0] {
		case types.QueryMatchResults:
			res, err = queryMatchResults(ctx, path[1:], req, keeper)
		case types.QueryDealList:
			res, err = queryDeals(ctx, path[1:], req, keeper)
		case types.QueryFeeDetails:
			res, err = queryFeeDetails(ctx, path[1:], req, keeper)
		case types.QueryOrderList:
			res, err = queryOrderList(ctx, path[1:], req, keeper)
		case types.QueryTxList:
			res, err = queryTxList(ctx, path[1:], req, keeper)
		case types.QueryCandleList:
			if keeper.Config.EnableMktCompute {
				res, err = queryCandleList(ctx, path[1:], req, keeper)
			} else {
				res, err = queryCandleListFromMarketKeeper(ctx, path[1:], req, keeper)
			}
		case types.QueryTickerList:
			if keeper.Config.EnableMktCompute {
				res, err = queryTickerList(ctx, path[1:], req, keeper)
			} else {
				res, err = queryTickerListFromMarketKeeper(ctx, path[1:], req, keeper)
			}

		case types.QueryTickerListV2:
			if keeper.Config.EnableMktCompute {
				res, err = queryTickerListV2(ctx, path[1:], req, keeper)
			} else {
				res, err = queryTickerListFromMarketKeeperV2(ctx, path[1:], req, keeper)
			}
		case types.QueryTickerV2:
			if keeper.Config.EnableMktCompute {
				res, err = queryTickerV2(ctx, path[1:], req, keeper)
			} else {
				res, err = queryTickerFromMarketKeeperV2(ctx, path[1:], req, keeper)
			}
		case types.QueryInstrumentsV2:
			res, err = queryInstrumentsV2(ctx, path[1:], req, keeper)
		case types.QueryOrderListV2:
			res, err = queryOrderListV2(ctx, path[1:], req, keeper)
		case types.QueryOrderV2:
			res, err = queryOrderV2(ctx, path[1:], req, keeper)
		case types.QueryCandleListV2:
			if keeper.Config.EnableMktCompute {
				res, err = queryCandleListV2(ctx, path[1:], req, keeper)
			} else {
				res, err = queryCandleListFromMarketKeeperV2(ctx, path[1:], req, keeper)
			}
		case types.QueryMatchResultsV2:
			res, err = queryMatchResultsV2(ctx, path[1:], req, keeper)
		case types.QueryFeeDetailsV2:
			res, err = queryFeeDetailsV2(ctx, path[1:], req, keeper)
		case types.QueryDealListV2:
			res, err = queryDealsV2(ctx, path[1:], req, keeper)
		case types.QueryTxListV2:
			res, err = queryTxListV2(ctx, path[1:], req, keeper)
		default:
			res, err = nil, sdk.ErrUnknownRequest("unknown backend endpoint")
		}

		if err != nil {
			response := common.GetErrorResponse(-1, "", err.Error())
			res, eJson := json.Marshal(response)
			if eJson != nil {
				return nil, sdk.ErrInternal(eJson.Error())
			}
			return res, err
		}

		return res, nil
	}
}

// nolint: unparam
func queryDeals(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryDealsParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	if params.Side != "" && params.Side != orderTypes.BuyOrder && params.Side != orderTypes.SellOrder {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("Side should not be %s", params.Side))
	}

	offset, limit := common.GetPage(params.Page, params.PerPage)
	deals, total := keeper.GetDeals(ctx, params.Address, params.Product, params.Side, params.Start, params.End, offset, limit)
	var response *common.ListResponse
	if len(deals) > 0 {
		response = common.GetListResponse(total, params.Page, params.PerPage, deals)
	} else {
		response = common.GetEmptyListResponse(total, params.Page, params.PerPage)
	}
	bz, err := json.Marshal(response)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}
	return bz, nil
}

// nolint: unparam
func queryMatchResults(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryMatchParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	offset, limit := common.GetPage(params.Page, params.PerPage)
	matches, total := keeper.GetMatchResults(ctx, params.Product, params.Start, params.End, offset, limit)
	var response *common.ListResponse
	if len(matches) > 0 {
		response = common.GetListResponse(total, params.Page, params.PerPage, matches)
	} else {
		response = common.GetEmptyListResponse(total, params.Page, params.PerPage)
	}
	bz, err := json.Marshal(response)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}
	return bz, nil
}

// nolint: unparam
func queryFeeDetails(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryFeeDetailsParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}
	_, err = sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("invalid address", err.Error()))
	}

	offset, limit := common.GetPage(params.Page, params.PerPage)
	feeDetails, total := keeper.GetFeeDetails(ctx, params.Address, offset, limit)
	var response *common.ListResponse
	if len(feeDetails) > 0 {
		response = common.GetListResponse(total, params.Page, params.PerPage, feeDetails)
	} else {
		response = common.GetEmptyListResponse(total, params.Page, params.PerPage)
	}
	bz, err := json.Marshal(response)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}
	return bz, nil
}

func queryCandleList(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {

	var params types.QueryKlinesParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}
	ctx.Logger().Debug(fmt.Sprintf("queryCandleList : %+v", params))
	restData, err := keeper.GetCandles(params.Product, params.Granularity, params.Size)

	var response *common.BaseResponse
	if err != nil {
		response = common.GetErrorResponse(-1, "", err.Error())
	} else {
		response = common.GetBaseResponse(restData)
	}

	bz, err := json.Marshal(response)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}
	return bz, nil
}

func queryCandleListFromMarketKeeper(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {

	var params types.QueryKlinesParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}
	ctx.Logger().Debug(fmt.Sprintf("queryCandleList : %+v", params))
	// should init token pair map here
	keeper.marketKeeper.InitTokenPairMap(ctx, keeper.dexKeeper)
	restData, err := keeper.getCandlesByMarketKeeper(params.Product, params.Granularity, params.Size)

	var response *common.BaseResponse
	if err != nil {
		response = common.GetErrorResponse(-1, "", err.Error())
	} else {
		response = common.GetBaseResponse(restData)
	}

	bz, err := json.Marshal(response)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}
	return bz, nil
}

func queryTickerList(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	params := types.QueryTickerParams{}
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data, ", err.Error()))
	}

	products := []string{}
	if params.Product != "" {
		products = append(products, params.Product)
	} else {
		products = keeper.GetAllProducts(ctx)
	}

	addedTickers := []types.Ticker{}
	tickers := keeper.GetTickers(products, params.Count)
	for _, p := range products {

		exists := false
		for _, t := range tickers {
			if p == t.Product {
				exists = true
				break
			}
		}

		if !exists {
			//tmpPrice := keeper.orderKeeper.GetLastPrice(ctx, p)
			tmpTicker := types.Ticker{
				Price:            -1,
				Product:          p,
				Symbol:           p,
				Open:             0,
				Close:            0,
				High:             0,
				Low:              0,
				Volume:           0,
				Change:           0,
				ChangePercentage: "0.00%",
				Timestamp:        time.Now().Unix(),
			}
			addedTickers = append(addedTickers, tmpTicker)
		}

	}

	if len(addedTickers) > 0 {
		tickers = append(tickers, addedTickers...)
	}

	var sortedTickers types.Tickers = tickers
	sort.Sort(sortedTickers)

	response := common.GetBaseResponse(sortedTickers)
	bz, err := json.Marshal(response)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}
	return bz, nil
}

func queryTickerListFromMarketKeeper(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {

	params := types.QueryTickerParams{}
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data, ", err.Error()))
	}
	keeper.marketKeeper.InitTokenPairMap(ctx, keeper.dexKeeper)
	tickers, err := keeper.marketKeeper.GetTickers()

	var products []string
	if params.Product != "" {
		products = append(products, params.Product)
	} else {
		products = keeper.GetAllProducts(ctx)
	}

	var addedTickers []map[string]string
	for _, p := range products {

		exists := false
		for _, t := range tickers {
			if p == t["product"] {
				exists = true
				break
			}
		}

		if !exists {
			tmpTicker := map[string]string{
				"price":   "-1",
				"product": p,
				"symbol":  p,
				"open":    "0",
				"close":   "0",
				"high":    "0",
				"low":     "0",
				"volume":  "0",
				"change":  "0",
				//"changePercentage": "0.00%",
				"timestamp": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			}
			addedTickers = append(addedTickers, tmpTicker)
		}

	}

	if len(addedTickers) > 0 {
		tickers = append(tickers, addedTickers...)
	}

	if len(tickers) > params.Count {
		tickers = tickers[0:params.Count]
	}

	var response *common.BaseResponse
	if err != nil {
		response = common.GetErrorResponse(-1, "", err.Error())
	} else {
		response = common.GetBaseResponse(tickers)
	}

	bz, err := json.Marshal(response)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}
	return bz, nil
}

func queryOrderList(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	isOpen := path[0] == "open"
	var params types.QueryOrderListParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}
	_, err = sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("invalid address", err.Error()))
	}
	offset, limit := common.GetPage(params.Page, params.PerPage)
	orders, total := keeper.GetOrderList(ctx, params.Address, params.Product, params.Side, isOpen,
		offset, limit, params.Start, params.End, params.HideNoFill)

	var response *common.ListResponse
	if len(orders) > 0 {
		response = common.GetListResponse(total, params.Page, params.PerPage, orders)
	} else {
		response = common.GetEmptyListResponse(total, params.Page, params.PerPage)
	}
	bz, err := json.Marshal(response)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}

	return bz, nil
}

func queryTxList(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryTxListParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}
	_, err = sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("invalid address", err.Error()))
	}
	offset, limit := common.GetPage(params.Page, params.PerPage)
	txs, total := keeper.GetTransactionList(ctx, params.Address, params.TxType, params.StartTime, params.EndTime, offset, limit)

	var response *common.ListResponse
	if len(txs) > 0 {
		response = common.GetListResponse(total, params.Page, params.PerPage, txs)
	} else {
		response = common.GetEmptyListResponse(total, params.Page, params.PerPage)
	}
	bz, err := json.Marshal(response)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}
	return bz, nil
}
