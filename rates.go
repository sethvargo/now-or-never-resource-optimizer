package main

import "fmt"

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

// RateTable represents a fixed collection of exchange rates. The game has a
// default rate table, but player modifications may mutate the rate table. For
// example, there are cards that change the rate at which certain resources are
// traded.
type RateTable []*ExchangeRate

// defaultRateTable is the default rate table.
var defaultRateTable RateTable = []*ExchangeRate{
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
}
