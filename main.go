package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/sethvargo/optim/rates"
)

func main() {
	cache := make(rates.Cache, 8)
	rateTable := rates.DefaultRateTable()

	js.Global().Set("bestTrade", jsBestTrade(cache, rateTable))
	<-make(chan struct{})
}

func jsError(err error) js.Value {
	return js.Global().Get("Error").New(err.Error())
}

type JSTrade struct {
	*rates.ResourceAlloc

	ShellModifier bool `json:"sm"`
	ToolModifier  bool `json:"tm"`
}

type JSResponse struct {
	BestTrade *rates.Trade     `json:"t"`
	RateTable *rates.RateTable `json:"r"`
}

func jsBestTrade(cache rates.Cache, rateTable *rates.RateTable) js.Func {
	return js.FuncOf(func(this js.Value, parentArgs []js.Value) interface{} {
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve, reject := args[0], args[1]

			go func() {
				if len(parentArgs) != 1 {
					reject.Invoke(jsError(fmt.Errorf("invalid input")))
					return
				}

				in := parentArgs[0].String()

				var trade JSTrade
				if err := json.Unmarshal([]byte(in), &trade); err != nil {
					reject.Invoke(jsError(fmt.Errorf("failed to decode json: %w", err)))
					return
				}

				// Modify rate table
				if rateTable.SetShellModifier(trade.ShellModifier) {
					cache.Invalidate()
				}
				if rateTable.SetToolModifier(trade.ToolModifier) {
					cache.Invalidate()
				}

				trades := rates.ExchangeWith(cache, rateTable, trade.ResourceAlloc)
				if len(trades) < 1 {
					reject.Invoke(jsError(fmt.Errorf("no results")))
					return
				}

				b, err := json.Marshal(&JSResponse{
					BestTrade: trades[0],
					RateTable: rateTable,
				})
				if err != nil {
					reject.Invoke(jsError(fmt.Errorf("failed to create json: %w", err)))
					return
				}
				resolve.Invoke(js.ValueOf(string(b)))
			}()

			return nil
		})

		jsPromise := js.Global().Get("Promise")
		return jsPromise.New(handler)
	})
}
