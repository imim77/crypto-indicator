package main

import (
	"fmt"
	"log"
	"strconv"

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

func render() {
	w, h := termbox.Size()
	for i := 0; i < w; i++ {
		termbox.SetCell(i, 0, '-', termbox.ColorMagenta, termbox.ColorDefault)
		termbox.SetCell(i, h-1, '-', termbox.ColorMagenta, termbox.ColorDefault)
	}

	for i := 0; i < h; i++ {
		termbox.SetCell(0, i, '|', termbox.ColorMagenta, termbox.ColorDefault)
		termbox.SetCell(w-1, i, '|', termbox.ColorMagenta, termbox.ColorDefault)
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

		if entry, ok := ob.Asks.Get(getAskByPrice(price)); ok {
			if volume == 0 {
				ob.Asks.Delete(entry)
			} else {
				entry.Volume = volume
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

		if entry, ok := ob.Bids.Get(getBidByPrice(price)); ok {
			if volume == 0 {
				ob.Bids.Delete(entry)
			} else {
				entry.Volume = volume
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

func (ob *Orderbook) getBids() []*OrderbookEntry {
	var (
		depth = 10
		bids  = make([]*OrderbookEntry, depth)
		it    = ob.Bids.Iterator(nil, nil)
		i     = 0
	)
	for it.Next() {
		if i == depth {
			break
		}
		bids[i] = it.Item()
		i++
	}
	it.Release()
	return bids
}

func (ob *Orderbook) getAsks() []*OrderbookEntry {
	var (
		depth = 10
		asks  = make([]*OrderbookEntry, depth)
		it    = ob.Asks.Iterator(nil, nil)
		i     = 0
	)
	for it.Next() {
		if i == depth {
			break
		}
		asks[i] = it.Item()
		i++
	}
	it.Release()
	return asks
}

func (ob *Orderbook) render(x, y int) {
	for i, ask := range ob.getAsks() {
		if ask == nil {
			continue
		}
		price := fmt.Sprintf("%.2f", ask.Price)
		volume := fmt.Sprintf("%.2f", ask.Price)
		renderText(x, y+i, price, termbox.ColorRed)
		renderText(x+10, y+i, volume, termbox.ColorCyan)

	}
	for i, bid := range ob.getBids() {
		if bid == nil {
			continue
		}
		price := fmt.Sprintf("%.2f", bid.Price)
		volume := fmt.Sprintf("%.2f", bid.Price)
		renderText(x, 10+i, price, termbox.ColorGreen)
		renderText(x+10, 10+i, volume, termbox.ColorCyan)

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
	eventch := make(chan termbox.Event, 1)
	go func() {
		for {
			eventch <- termbox.PollEvent()
		}
	}()
	for is_running {
		select {
		case event := <-eventch:
			switch event.Key {
			case termbox.KeyEsc:
				is_running = false
			}
		default:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			ob.render(2, 2)
			render()
			//time.Sleep(time.Millisecond * 16)
			termbox.Flush()
		}
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
	}
}
