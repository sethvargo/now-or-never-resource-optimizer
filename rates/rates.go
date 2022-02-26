package rates

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ResourceAlloc represents a resource allocation requirement or constraint.
type ResourceAlloc struct {
	// S is the number of shells.
	S uint8 `json:"s,omitempty"`

	// T is the number of tools.
	T uint8 `json:"t,omitempty"`

	// D is the number of demons.
	D uint8 `json:"d,omitempty"`

	// C is the number of crystals.
	C uint8 `json:"c,omitempty"`
}

// Hash returns a hash for the allocation. All allocations with the same
// resource constraints hash to the same value.
// TODO(sethvargo): make more efficient
func (r *ResourceAlloc) Hash() string {
	return fmt.Sprintf("s:%d t:%d d:%d c:%d", r.S, r.T, r.D, r.C)
}

// HashCode returns a 32-bit integer which is guaranteed to unique represent a
// ResourceAlloc.
func (r *ResourceAlloc) HashCode() uint32 {
	return uint32(r.S)*1_000_000 + uint32(r.T)*10_000 + uint32(r.D)*100 + uint32(r.C)
}

// IsEmpty returns true if the resource allocation has no remaining resources
// (all resources are at zero), or false otherwise.
func (r *ResourceAlloc) IsEmpty() bool {
	return r.S == 0 && r.T == 0 && r.D == 0 && r.C == 0
}

// Sub attempts to make a trade with the provided input. If the trade is
// possible, it returns a new resource allocation with the resources removed. If
// the trade is impossible, it returns false. In all cases, the current resource
// allocation is never modified.
func (r *ResourceAlloc) Sub(trade *ResourceAlloc) (*ResourceAlloc, bool) {
	if r.S < trade.S || r.T < trade.T || r.D < trade.D || r.C < trade.C {
		return nil, false
	}

	return &ResourceAlloc{
		S: r.S - trade.S,
		T: r.T - trade.T,
		D: r.D - trade.D,
		C: r.C - trade.C,
	}, true
}

// ResourceAllocs is a collection of resource allocations. It implements sort.
type ResourceAllocs []*ResourceAlloc

func (r ResourceAllocs) Len() int { return len(r) }

func (r ResourceAllocs) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

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

// Trade represents an allocation for value mapping.
type Trade struct {
	R []*ResourceAlloc
	V uint8
}

// ExchangeRate represents an exchange of resources for their value.
type ExchangeRate struct {
	// R is the resource requirement for this exchange.
	R *ResourceAlloc `json:"r"`

	// V is the value for redeeming the resources in the allocation.
	V uint8 `json:"v"`
}

// RateTable represents the exchange rates.
type RateTable struct {
	rates []*ExchangeRate

	shellModifier bool
	toolModifier  bool
}

// MarshalJSON is a custom JSON marshaller for the rate table. It returns a map
// of the resource constraint (as JSON) to the value.
func (r *RateTable) MarshalJSON() ([]byte, error) {
	m := make(map[string]uint8, len(r.rates))
	for _, v := range r.rates {
		h, err := json.Marshal(v.R)
		if err != nil {
			return nil, err
		}
		m[string(h)] = v.V
	}

	return json.Marshal(m)
}

// defaultRateTable is the default rate table.
var defaultRateTable *RateTable = &RateTable{
	rates: []*ExchangeRate{
		{
			R: &ResourceAlloc{D: 2, C: 2},
			V: 14,
		},
		{
			R: &ResourceAlloc{S: 1, T: 1, D: 1, C: 1},
			V: 12,
		},
		{
			R: &ResourceAlloc{C: 3},
			V: 11,
		},
		{
			R: &ResourceAlloc{S: 1, T: 1, D: 1},
			V: 9,
		},
		{
			R: &ResourceAlloc{T: 1, D: 2},
			V: 9,
		},
		{
			R: &ResourceAlloc{S: 1, T: 2},
			V: 7,
		},
		{
			R: &ResourceAlloc{S: 3},
			V: 5,
		},
		{
			R: &ResourceAlloc{C: 1},
			V: 2,
		},
		{
			R: &ResourceAlloc{D: 1},
			V: 2,
		},
		{
			R: &ResourceAlloc{T: 1},
			V: 2,
		},
		{
			R: &ResourceAlloc{S: 1},
			V: 1,
		},
	},
}

// DefaultRateTable returns a copy of the default rate table.
func DefaultRateTable() *RateTable {
	rateTable := &RateTable{
		rates: make([]*ExchangeRate, len(defaultRateTable.rates)),
	}
	for i, v := range defaultRateTable.rates {
		rateTable.rates[i] = &ExchangeRate{
			R: &ResourceAlloc{
				S: v.R.S,
				T: v.R.T,
				D: v.R.D,
				C: v.R.C,
			},
			V: v.V,
		}
	}
	return rateTable
}

// SetShellModifier enables the shell modifier to the rate table.
func (r *RateTable) SetShellModifier(enable bool) bool {
	if enable {
		if r.shellModifier {
			// Already enabled
			return false
		}

		// Enable it
		r.shellModifier = true
		r.rates[6].V = 8
		return true
	}

	// Enabled but shouldn't be anymore
	if r.shellModifier {
		r.shellModifier = false
		r.rates[6].V = 5
		return true
	}

	// Not enabled and shouldn't be
	return false
}

// SetToolModifier enables the tool modification rate table.
func (r *RateTable) SetToolModifier(enable bool) bool {
	if enable {
		if r.toolModifier {
			// Already enabled
			return false
		}

		// Enable it
		r.toolModifier = true
		r.rates = append(r.rates, &ExchangeRate{
			R: &ResourceAlloc{T: 3},
			V: 10,
		})
		return true
	}

	// Enabled but shouldn't be anymore
	if r.toolModifier {
		r.toolModifier = false
		r.rates = (r.rates)[:11]
		return true
	}

	// Not enabled and shouldn't be
	return false
}

// Cache represents a resource cache.
type Cache map[uint32][]*Trade

// Invalidate purges the cache.
func (c Cache) Invalidate() {
	for k := range c {
		delete(c, k)
	}
}

// Exchange does a default exchange with the default rate table and cache.
func Exchange(hand *ResourceAlloc) []*Trade {
	cache := make(map[uint32][]*Trade, 8)
	return ExchangeWith(cache, DefaultRateTable(), hand)
}

// ExchangeWith calculates the exchange rate with the given rate table.
func ExchangeWith(cache Cache, rateTable *RateTable, hand *ResourceAlloc) []*Trade {
	parent := make([]*Trade, 0)

	if hand == nil {
		return nil
	}

	// Check if we have already computed the optimal exchange here.
	cached, ok := cache[hand.HashCode()]
	if ok {
		return cached
	}

	for _, rate := range rateTable.rates {
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

		for _, trade := range ExchangeWith(cache, rateTable, remaining) {
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
