package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anakinrm/crypto-exchange/server"
)

const Endpoint = "http://localhost:3000"

type PlaceOrderParams struct {
	UserID int64
	Bid    bool
	// Price only needed for placing LIMIT orders
	Price float64
	Size  float64
}

type Client struct {
	*http.Client
}

func NewClient() *Client {
	return &Client{
		Client: http.DefaultClient,
	}
}

func (c *Client) GetOrders(userID int64) ([]server.Order, error) {
	e := fmt.Sprintf("%s/order/%d", Endpoint, userID)
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	orders := []server.Order{}
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (c *Client) PlaceMarketOrder(p *PlaceOrderParams) (*server.PlaceOrderResponse, error) {
	params := &server.PlaceOrderRequest{
		UserID: p.UserID,
		Type:   server.MarketOrder,
		Bid:    p.Bid,
		Size:   p.Size,
		Market: server.MarketETH,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	e := Endpoint + "/order"
	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderResponse := &server.PlaceOrderResponse{}
	if err := json.NewDecoder(resp.Body).Decode(placeOrderResponse); err != nil {
		return nil, err
	}

	return placeOrderResponse, nil
}

// get Best Price from the server
const (
	bestBidPrice bestPriceType = "bestbid"
	bestAskPrice bestPriceType = "bestask"
)

type bestPriceType string

func (c *Client) getBestPrice(priceType bestPriceType) (float64, error) {

	e := fmt.Sprintf("%s/book/ETH/%s", Endpoint, priceType)
	req, err := http.NewRequest(http.MethodGet, e, nil)
	if err != nil {
		return 0.0, err
	}

	resp, err := c.Do(req)
	if err != err {
		return 0.0, err
	}

	priceResp := &server.PriceResponse{}
	if err := json.NewDecoder(resp.Body).Decode(priceResp); err != nil {
		return 0.0, err
	}

	return priceResp.Price, err
}

func (c *Client) GetBestBidPrice() (float64, error) {
	return c.getBestPrice(bestBidPrice)
}

func (c *Client) GetBestAskPrice() (float64, error) {
	return c.getBestPrice(bestAskPrice)
}

func (c *Client) CancelOrder(orderID int64) error {
	e := fmt.Sprintf("%s/order/%d", Endpoint, orderID)
	req, err := http.NewRequest(http.MethodDelete, e, nil)
	if err != nil {
		return err
	}

	_, err = c.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) PlaceLimitOrder(p *PlaceOrderParams) (*server.PlaceOrderResponse, error) {
	params := &server.PlaceOrderRequest{
		UserID: p.UserID,
		Type:   server.LimitOrder,
		Bid:    p.Bid,
		Size:   p.Size,
		Price:  p.Price,
		Market: server.MarketETH,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	e := Endpoint + "/order"
	req, err := http.NewRequest(http.MethodPost, e, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderResponse := &server.PlaceOrderResponse{}
	if err := json.NewDecoder(resp.Body).Decode(placeOrderResponse); err != nil {
		return nil, err
	}

	return placeOrderResponse, nil
}
