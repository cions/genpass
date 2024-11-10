// Copyright (c) 2024 cions
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
	Lo, Hi rune
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
	if a.Hi < b {
		return -1
	}
	if a.Lo > b {
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
		lo = set.ranges[i].Lo
	}
	if found2 {
		hi = set.ranges[j].Hi
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
		for j < len(set.ranges) && cur.Hi+1 == set.ranges[j].Lo {
			cur.Hi = set.ranges[j].Hi
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
		size += int64(r.Hi) - int64(r.Lo) + 1
		cumsizes[i] = size
	}
	return &Picker{set.ranges, cumsizes, size}
}

func (set *RuneSet) String() string {
	var b strings.Builder
	for _, r := range set.ranges {
		b.WriteRune(r.Lo)
		b.WriteRune('-')
		b.WriteRune(r.Hi)
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
	ri, found := slices.BinarySearch(p.cumSizes, i)
	if found {
		ri++
	}
	offset := i
	if ri > 0 {
		offset -= p.cumSizes[ri-1]
	}
	return p.ranges[ri].Lo + rune(offset)
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
