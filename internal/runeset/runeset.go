// Copyright (c) 2024-2025 cions
// Licensed under the MIT License. See LICENSE for details.

package runeset

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"unicode"
)

type Range struct {
	lo, hi rune
}

type RuneSet struct {
	ranges []Range
}

type Picker struct {
	ranges   []Range
	cumSizes []int64
	size     int64
}

func compare(a Range, b rune) int {
	if a.hi < b {
		return -1
	}
	if a.lo > b {
		return 1
	}
	return 0
}

func (set *RuneSet) Add(r rune) {
	i, found := slices.BinarySearchFunc(set.ranges, r, compare)
	if !found {
		set.ranges = slices.Insert(set.ranges, i, Range{r, r})
	}
}

func (set *RuneSet) AddRange(lo, hi rune) {
	if lo > hi {
		panic("runeset: lo must be smaller than or equals to hi")
	}
	i, found1 := slices.BinarySearchFunc(set.ranges, lo, compare)
	j, found2 := slices.BinarySearchFunc(set.ranges, hi, compare)
	if found1 {
		lo = set.ranges[i].lo
	}
	if found2 {
		hi = set.ranges[j].hi
		j++
	}
	set.ranges = slices.Replace(set.ranges, i, j, Range{lo, hi})
}

func (set *RuneSet) AddRangeTable(table *unicode.RangeTable) {
	for _, r := range table.R16 {
		if r.Stride == 1 {
			set.AddRange(rune(r.Lo), rune(r.Hi))
		} else {
			for x := r.Lo; x <= r.Hi; x += r.Stride {
				set.Add(rune(x))
			}
		}
	}
	for _, r := range table.R32 {
		if r.Stride == 1 {
			set.AddRange(rune(r.Lo), rune(r.Hi))
		} else {
			for x := r.Lo; x <= r.Hi; x += r.Stride {
				set.Add(rune(x))
			}
		}
	}
}

func (set *RuneSet) MergeAdjacents() {
	i, j := 0, 0
	for j < len(set.ranges) {
		cur := set.ranges[j]
		j++
		for j < len(set.ranges) && cur.hi+1 == set.ranges[j].lo {
			cur.hi = set.ranges[j].hi
			j++
		}
		set.ranges[i] = cur
		i++
	}
	set.ranges = set.ranges[:i]
}

func (set *RuneSet) Picker() *Picker {
	var size int64
	cumsizes := make([]int64, len(set.ranges))
	for i, r := range set.ranges {
		size += int64(r.hi) - int64(r.lo) + 1
		cumsizes[i] = size
	}
	return &Picker{set.ranges, cumsizes, size}
}

func (set *RuneSet) String() string {
	var b strings.Builder
	for _, r := range set.ranges {
		b.WriteRune(r.lo)
		b.WriteRune('-')
		b.WriteRune(r.hi)
	}
	return b.String()
}

func (p *Picker) Size() int64 {
	return p.size
}

func (p *Picker) Get(i int64) rune {
	if i < 0 || i >= p.size {
		panic("runeset: out of bounds")
	}
	ridx, found := slices.BinarySearch(p.cumSizes, i)
	if found {
		ridx++
	}
	offset := i
	if ridx > 0 {
		offset -= p.cumSizes[ridx-1]
	}
	return p.ranges[ridx].lo + rune(offset)
}

func (p *Picker) Random() rune {
	n := big.NewInt(p.size)
	i, err := rand.Int(rand.Reader, n)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand: %v", err))
	} else if !i.IsInt64() {
		panic("crypto/rand: out of range")
	}
	return p.Get(i.Int64())
}
