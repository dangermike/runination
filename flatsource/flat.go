package flatsource

import (
	"unicode"
)

type CharSourceFlat []rune

func New(tables []*unicode.RangeTable) CharSourceFlat {
	size := 0
	for _, table := range tables {
		for _, r := range table.R16 {
			size += (int((r.Hi-r.Lo)/r.Stride) + 1) // +1 because the range is inclusive
		}

		for _, r := range table.R32 {
			size += (int((r.Hi-r.Lo)/r.Stride) + 1) // +1 because the range is inclusive
		}
	}

	cs := make(CharSourceFlat, 0, size)

	for _, table := range tables {
		for _, rng := range table.R16 {
			// if you don't up-convert this, you can get caught with
			// wrapping before the condition is checked
			lo, hi, stride := int32(rng.Lo), int32(rng.Hi), int32(rng.Stride)
			for r := lo; r <= hi; r += stride {
				cs = append(cs, rune(r))
			}
		}
		for _, rng := range table.R32 {
			for r := rng.Lo; r <= rng.Hi; r += rng.Stride {
				cs = append(cs, rune(r))
			}
		}
	}

	return cs
}

func (cs CharSourceFlat) At(num int) rune {
	return cs[num]
}

func (cs CharSourceFlat) Count() int {
	return len(cs)
}
