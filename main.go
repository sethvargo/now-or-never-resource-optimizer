package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"syscall/js"
)

func main() {
	js.Global().Set("rateTable", jsRateTable())
	js.Global().Set("bestTrade", jsBestTrade2())
	<-make(chan struct{})
}

func jsError(err error) js.Value {
	return js.Global().Get("Error").New(err.Error())
}

func jsRateTable() js.Func {
	return js.FuncOf(func(this js.Value, parentArgs []js.Value) interface{} {
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve, reject := args[0], args[1]

			go func() {
				m := make(map[string]uint8, len(defaultRateTable))
				for _, v := range defaultRateTable {
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

func jsBestTrade2() js.Func {
	return js.FuncOf(func(this js.Value, parentArgs []js.Value) interface{} {
		handler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resolve, reject := args[0], args[1]

			go func() {
				if len(parentArgs) != 1 {
					reject.Invoke(jsError(fmt.Errorf("invalid input")))
					return
				}

				in := parentArgs[0].String()

				var hand ResourceAlloc
				if err := json.Unmarshal([]byte(in), &hand); err != nil {
					reject.Invoke(jsError(fmt.Errorf("failed to decode json: %w", err)))
					return
				}

				result := exchange(defaultRateTable, &hand)
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

var cache = make(map[uint32][]*Trade, 128)

func exchange(rateTable RateTable, hand *ResourceAlloc) []*Trade {
	// Check if we have already computed the optimal exchange here.
	cached, ok := cache[hand.HashCode()]
	if ok {
		return cached
	}

	parent := make([]*Trade, 0)

	for _, rate := range rateTable {
		remaining, ok := hand.Sub(rate.R)
		if !ok {
			continue
		}

		// If the trade succeed, but the hand is empty, we're done
		if remaining.IsEmpty() {
			child := make([]*ResourceAlloc, 0)
			child = append(child, rate.R)
			parent = append(parent, &Trade{R: child, V: rate.V})
			continue
		}

		for _, trade := range exchange(rateTable, remaining) {
			child := make([]*ResourceAlloc, 0)
			child = append(child, rate.R)
			child = append(child, trade.R...)
			parent = append(parent, &Trade{R: child, V: rate.V + trade.V})
		}
	}

	// Remove any "duplicate" answers. It's possible that A,B,C and C,A,B were
	// both permuted as possible trades, but they are equivalent.
	m := make(map[string]struct{}, len(parent))
	i := 0
	for _, v := range parent {
		h := ResourceAllocs(v.R).Hash()
		if _, ok := m[h]; !ok {
			m[h] = struct{}{}
			parent[i] = v
			i++
		}
	}
	for j := i; j < len(parent); j++ {
		parent[j] = nil
	}
	parent = parent[:i]
	sort.Slice(parent, func(i, j int) bool {
		if parent[j].V == parent[i].V {
			if len(parent[j].R) == len(parent[i].R) {
				// If we got this far, just trust the insertion order since it's also
				// consistent.
				return false
			}
			return len(parent[j].R) < len(parent[i].R)
		}
		return parent[j].V < parent[i].V
	})

	// Cache the highest-scoring trades
	highest := make([]*Trade, 0, 2)
	if len(parent) > 0 {
		highest = append(highest, parent[0])
		for i := 1; i < len(parent); i++ {
			if parent[i].V == highest[0].V {
				highest = append(highest, parent[i])
				continue
			}
			break
		}
	}
	cache[hand.HashCode()] = highest

	return parent
}

type ResourceAllocs []*ResourceAlloc

func (r ResourceAllocs) Len() int {
	return len(r)
}

func (r ResourceAllocs) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ResourceAllocs) Less(i, j int) bool {
	a := uint32(r[i].S)*1_000_000 + uint32(r[i].T)*10_000 + uint32(r[i].D)*100 + uint32(r[i].C)
	b := uint32(r[j].S)*1_000_000 + uint32(r[j].T)*10_000 + uint32(r[j].D)*100 + uint32(r[j].C)
	return b < a
}

func (r ResourceAllocs) Hash() string {
	sort.Sort(r)

	var b strings.Builder
	for i, v := range r {
		if i > 0 {
			b.WriteString("|")
		}
		b.WriteString(v.Hash())
	}
	return b.String()
}
