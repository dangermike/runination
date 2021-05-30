package rangedsource

import (
	"sort"
	"unicode"
)

type CharSourceRanged struct {
	unicode.RangeTable
	rangeEnds   []int
	rangeStarts []int
	count       int
}

func New(tables []*unicode.RangeTable) *CharSourceRanged {
	cnt16 := 0
	cnt32 := 0
	for _, t := range tables {
		cnt16 += len(t.R16)
		cnt32 += len(t.R32)
	}
	cs := CharSourceRanged{}
	cs.R16 = make([]unicode.Range16, 0, cnt16)
	cs.R32 = make([]unicode.Range32, 0, cnt32)
	cs.rangeEnds = make([]int, cnt16+cnt32)
	cs.rangeStarts = make([]int, cnt16+cnt32)

	for _, t := range tables {
		cs.R16 = append(cs.R16, t.R16...)
		cs.R32 = append(cs.R32, t.R32...)
	}

	for i, r := range cs.R16 {
		cs.rangeStarts[i] = cs.count
		cs.count += (int((r.Hi-r.Lo)/r.Stride) + 1) // +1 because the range is inclusive
		cs.rangeEnds[i] = cs.count - 1
	}

	for i, r := range cs.R32 {
		cs.rangeStarts[i+cnt16] = cs.count
		cs.count += (int((r.Hi-r.Lo)/r.Stride) + 1) // +1 because the range is inclusive
		cs.rangeEnds[i+cnt16] = cs.count - 1
	}

	return &cs
}

func (cs *CharSourceRanged) At(num int) rune {
	rangeIx := sort.SearchInts(cs.rangeEnds, num)
	charIx := num - cs.rangeStarts[rangeIx]
	if rangeIx >= len(cs.R16) {
		rangeIx -= len(cs.R16)
		return charAtNum32(cs.R32[rangeIx], uint32(charIx))
	}
	return charAtNum16(cs.R16[rangeIx], uint16(charIx))
}

func charAtNum16(r unicode.Range16, num uint16) rune {
	return rune(r.Lo + (r.Stride * num))
}

func charAtNum32(r unicode.Range32, num uint32) rune {
	return rune(r.Lo + (r.Stride * num))
}

func (cs *CharSourceRanged) Count() int {
	return cs.count
}
