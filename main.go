package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/VictorLowther/btree"
	"github.com/gorilla/websocket"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
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

// "novi main kao!?"
func main() {
	termbox.Init()
	conn, _, err := websocket.DefaultDialer.Dial(wsendpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	var (
		ob     = NewOrderBook()
		result BinanceDepthResponse
	)
	go func() {
		for {
			if err := conn.ReadJSON(&result); err != nil {
				log.Fatal(err)
			}
			ob.handleDepthResponse(result.Data)
		}
	}()
	is_running := true
	go func() {
		time.Sleep(time.Second * 10)
		is_running = false
	}()
	for is_running {
		/*
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeySpace:
				case termbox.KeyEsc:
					break loop
				}
			default:
			}
		*/
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		ob.render(0, 0)
		termbox.Flush()
	}
}

func renderText(x int, y int, msg string, color termbox.Attribute) {
	for _, chr := range msg {
		termbox.SetCell(x, y, chr, color, termbox.ColorDefault)
		w := runewidth.RuneWidth(chr)
		x += w

	}
}

func (ob *Orderbook) handleDepthResponse(res BinanceDepthResult) {
	for _, ask := range res.Asks {
		price, _ := strconv.ParseFloat(ask[0], 64)
		volume, _ := strconv.ParseFloat(ask[1], 64)
		if volume == 0 {
			if entry, ok := ob.Asks.Get(getAskByPrice(price)); ok {
				//log.Printf("-- deleting level %.2f", price)
				ob.Asks.Delete(entry)
			}
			continue
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
			if entry, ok := ob.Bids.Get(getBidByPrice(price)); ok {
				//log.Printf("-- deleting level %.2f", price)
				ob.Bids.Delete(entry)
			}
			continue

		}
		entry := &OrderbookEntry{
			Price:  price,
			Volume: volume,
		}
		ob.Bids.Insert(entry)
	}
}

func (ob *Orderbook) render(x, y int) {
	it := ob.Asks.Iterator(nil, nil)
	i := 0
	for it.Next() {
		//fmt.Printf("%+v\n", it.Item())
		item := it.Item()
		priceStr := fmt.Sprintf("%.2f", item.Price)
		renderText(x, y+i, priceStr, termbox.ColorRed)
		i++
	}
	it = ob.Bids.Iterator(nil, nil)
	i = 0
	x = x + 15
	for it.Next() {
		//fmt.Printf("%+v\n", it.Item())
		item := it.Item()
		priceStr := fmt.Sprintf("%.2f", item.Price)
		renderText(x, y+i, priceStr, termbox.ColorGreen)
		i++
	}
}

func getAskByPrice(price float64) btree.CompareAgainst[*OrderbookEntry] {
	return func(e *OrderbookEntry) int {
		switch {
		case e.Price < price:
			return -1
		case e.Price > price:
			return 1
		default:
			return 0
		}
	}
}

func getBidByPrice(price float64) btree.CompareAgainst[*OrderbookEntry] {
	return func(e *OrderbookEntry) int {
		switch {
		case e.Price > price:
			return -1
		case e.Price < price:
			return 1
		default:
			return 0
		}
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

func _main() {
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
