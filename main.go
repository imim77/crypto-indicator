package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/VictorLowther/btree"
	"github.com/gorilla/websocket"
)

const wsendpoint = "wss://fstream.binance.com/stream?streams=btcusdt@depth"

func byBestBid(a, b *OrderbookEntry) bool {
	return a.Price >= b.Price
}

func byBestAsk(a, b *OrderbookEntry) bool {
	return a.Price < b.Price
}

type OrderbookEntry struct {
	Price  float64
	Volume float64
}

type Orderbook struct {
	Asks *btree.Tree[*OrderbookEntry]
	Bids *btree.Tree[*OrderbookEntry]
}

func (ob *Orderbook) handleDepthResponse(res BinanceDepthResult) {
	for _, ask := range res.Asks {
		price, _ := strconv.ParseFloat(ask[0], 64)
		volume, _ := strconv.ParseFloat(ask[1], 64)
		if volume == 0 {
			return
		}
		entry := &OrderbookEntry{
			Price:  price,
			Volume: volume,
		}
		ob.Asks.Insert(entry)
	}
	for _, bid := range res.Bids {
		price, _ := strconv.ParseFloat(bid[0], 64)
		volume, _ := strconv.ParseFloat(bid[1], 64)
		if volume == 0 {
			return
		}
		entry := &OrderbookEntry{
			Price:  price,
			Volume: volume,
		}
		ob.Bids.Insert(entry)
	}
}

func NewOrderBook() *Orderbook {
	return &Orderbook{
		Asks: btree.New(byBestAsk),
		Bids: btree.New(byBestBid),
	}
}

type BinanceDepthResult struct {
	Asks [][]string `json:"a"`
	Bids [][]string `json:"b"`
}

type BinanceDepthResponse struct {
	Stream string             `json:"stream"`
	Data   BinanceDepthResult `json:"data"`
}

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(wsendpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	var (
		ob     = NewOrderBook()
		result BinanceDepthResponse
	)

	for {
		if err := conn.ReadJSON(&result); err != nil {
			log.Fatal(err)
		}
		ob.handleDepthResponse(result.Data)
		it := ob.Asks.Iterator(nil, nil)
		for it.Next() {
			fmt.Printf("%+v\n", it.Item())
		}
	}

}
