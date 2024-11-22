package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/anakinrm/crypto-exchange/client"
	"github.com/anakinrm/crypto-exchange/server"
)

// resycle the time tick

var (
	tick = 2 * time.Second
)

func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(5 * time.Second)
	for {

		trades, err := c.GetTrades("ETH")
		if err != nil {
			panic(err)
		}

		if len(trades) > 0 {
			fmt.Printf("Excahnge price ======> %v\n", trades[len(trades)-1].Price)
		}

		//Sell
		otherMarketSellOrder := &client.PlaceOrderParams{
			UserID: 1,
			Bid:    false,
			Size:   1000,
		}
		orderResp, err := c.PlaceMarketOrder(otherMarketSellOrder)
		if err != nil {
			log.Println(orderResp)
		}

		marketSellOrder := &client.PlaceOrderParams{
			UserID: 666,
			Bid:    false,
			Size:   100,
		}

		sellOrderResp, err := c.PlaceMarketOrder(marketSellOrder)
		if err != nil {
			log.Println(sellOrderResp)
		}

		//Buy
		marketBuyOrder := &client.PlaceOrderParams{
			UserID: 666,
			Bid:    true,
			Size:   100,
		}

		orderResp, err = c.PlaceMarketOrder(marketBuyOrder)
		if err != nil {
			log.Println(orderResp)
		}

		<-ticker.C
	}
}

const userID = 2

func makeMarketSimple(c *client.Client) {
	ticker := time.NewTicker(tick)

	for {

		orders, err := c.GetOrders(userID)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("=-----------------------------------")
		fmt.Printf("Get orders from the client %+v\n", orders)
		fmt.Println("=-----------------------------------")

		bestAsk, err := c.GetBestAskPrice()
		if err != nil {
			log.Println(err)
		}

		bestBid, err := c.GetBestBidPrice()
		if err != nil {
			log.Println(err)
		}

		spread := math.Abs(bestBid - bestAsk)
		fmt.Println("exchange spread", spread)

		if len(orders.Bids) < 3 {
			//place the bid
			bidLimit := &client.PlaceOrderParams{
				UserID: 2,
				Bid:    true,
				Price:  bestBid + 100,
				Size:   1000,
			}

			bidOrderResp, err := c.PlaceLimitOrder(bidLimit)
			if err != nil {
				log.Println(bidOrderResp)
			}

		}

		if len(orders.Asks) < 3 {
			//place the ask
			askLimit := &client.PlaceOrderParams{
				UserID: 2,
				Bid:    false,
				Price:  bestAsk - 100,
				Size:   1000,
			}

			askOrderResp, err := c.PlaceLimitOrder(askLimit)
			if err != nil {
				log.Println(askOrderResp)
			}

		}

		fmt.Println("best ask price", bestAsk)
		fmt.Println("best bid price", bestBid)

		//its a channel which will better than sleep
		<-ticker.C
	}
}

func seedMarket(c *client.Client) error {
	ask := &client.PlaceOrderParams{
		UserID: 1,
		Bid:    false,
		Price:  10_000,
		Size:   5_000_000,
	}

	bid := &client.PlaceOrderParams{
		UserID: 1,
		Bid:    true,
		Price:  9_000,
		Size:   5_000_000,
	}
	_, err := c.PlaceLimitOrder(ask)
	if err != nil {
		return err
	}

	_, err = c.PlaceLimitOrder(bid)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	go server.StartServer()
	time.Sleep(1 * time.Second)

	c := client.NewClient()

	if err := seedMarket(c); err != nil {
		panic(err)
	}

	go makeMarketSimple(c)

	time.Sleep(1 * time.Second)
	marketOrderPlacer(c)

	select {}
}
