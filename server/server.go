package server

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
)

const (
	MarketETH Market = "ETH"

	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"

	exchangePrivateKey = "6b93be18f885aa07271e5be6f9cf2db740a63a1b73a24778f7e597e4a1cbfbe9"
)

type (
	Market    string
	OrderType string

	PlaceOrderRequest struct {
		UserID int64
		Type   OrderType // limit or market
		Bid    bool
		Size   float64
		Price  float64
		Market Market
	}

	Order struct {
		UserID    int64
		ID        int64
		Price     float64
		Size      float64
		Bid       bool
		Timestamp int64
	}

	OrderbookData struct {
		TotalBidVolume float64
		TotalAskVolume float64
		Asks           []*Order
		Bids           []*Order
	}

	MatchedOrders struct {
		UserID int64
		Price  float64
		Size   float64
		ID     int64
	}

	APIError struct {
		Error string
	}
)

func StartServer() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	ex, err := NewExchange(exchangePrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	ex.registerUser("f9065c72318979b6a164ed215bffbceec4d5f90e752d3c8d1192c0475bc473f6", 7)
	ex.registerUser("d2fa31763861778a3e19f29da5127539f96908d0406f75d69bd1cc32934b2934", 8)
	ex.registerUser("5e8f0213af74ba333b924a0d1db3a7c295e0918ccd06c8d89c1cb9046cca3be4", 666)

	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.cancelOrder)

	e.GET("/trades/:market", ex.handleGetTrades)
	e.GET("/order/:userID", ex.handleGetOrders)
	e.GET("/book/:market/asks", ex.handleGetBook)
	e.GET("/book/:market", ex.handleGetBook)
	e.GET("/book/:market/bestbid", ex.handleGetBestBid)
	e.GET("/book/:market/bestask", ex.handleGetBestAsk)

	e.Start(":3000")

}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}
