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

	js.Global().Set("rateTable", jsRateTable(rateTable))
	js.Global().Set("bestTrade", jsBestTrade(cache, rateTable))
	<-make(chan struct{})
}

func jsError(err error) js.Value {
	return js.Global().Get("Error").New(err.Error())
}

func jsRateTable(rateTable []*rates.ExchangeRate) js.Func {
	return js.FuncOf(func(this js.Value, parentArgs []js.Value) interface{} {
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve, reject := args[0], args[1]

			go func() {
				m := make(map[string]uint8, len(rateTable))
				for _, v := range rateTable {
					h, err := json.Marshal(v.R)
					if err != nil {
						reject.Invoke(jsError(fmt.Errorf("failed to marshal %#v: %w", v.R, err)))
						return
					}
					m[string(h)] = v.V
				}

				b, err := json.Marshal(m)
				if err != nil {
					reject.Invoke(jsError(fmt.Errorf("failed to marshal default rate table: %w", err)))
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

func jsBestTrade(cache rates.Cache, rateTable []*rates.ExchangeRate) js.Func {
	return js.FuncOf(func(this js.Value, parentArgs []js.Value) interface{} {
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve, reject := args[0], args[1]

			go func() {
				if len(parentArgs) != 1 {
					reject.Invoke(jsError(fmt.Errorf("invalid input")))
					return
				}

				in := parentArgs[0].String()

				var hand rates.ResourceAlloc
				if err := json.Unmarshal([]byte(in), &hand); err != nil {
					reject.Invoke(jsError(fmt.Errorf("failed to decode json: %w", err)))
					return
				}

				result := rates.ExchangeWith(cache, rateTable, &hand)
				if len(result) < 1 {
					reject.Invoke(jsError(fmt.Errorf("no results")))
					return
				}

				b, err := json.Marshal(result)
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
