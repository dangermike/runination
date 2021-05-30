package testhelper

import (
	"testing"
	"unicode"

	"github.com/dangermike/runination/source"
	"github.com/stretchr/testify/require"
)

type none struct{}

var (
	empty    = none{}
	Datasets = []struct {
		Name   string
		Tables []*unicode.RangeTable
	}{
		{"ASCII_hex_digit", []*unicode.RangeTable{unicode.ASCII_Hex_Digit}},
		{"ASCII_Alphanumeric", []*unicode.RangeTable{{R16: []unicode.Range16{{48, 57, 1}, {65, 90, 1}, {97, 122, 1}}}}},
		{"Latin", []*unicode.RangeTable{unicode.Latin}},
		{"Symbol", []*unicode.RangeTable{unicode.Symbol}},
		{"GraphicRanges", unicode.GraphicRanges},
	}
)

func tablesToRuneSet(tables []*unicode.RangeTable) map[rune]none {
	size := 0
	for _, table := range tables {
		for _, r := range table.R16 {
			size += (int((r.Hi-r.Lo)/r.Stride) + 1) // +1 because the range is inclusive
		}

		for _, r := range table.R32 {
			size += (int((r.Hi-r.Lo)/r.Stride) + 1) // +1 because the range is inclusive
		}
	}
	allRunes := make(map[rune]none, size)
	for _, table := range tables {
		for _, rng := range table.R16 {
			// test note: if you don't up-convert this, you can get caught with
			//  wrapping before the condition is checked
			lo, hi, stride := int32(rng.Lo), int32(rng.Hi), int32(rng.Stride)
			for r := lo; r <= hi; r += stride {
				allRunes[rune(r)] = empty
			}
		}
		for _, rng := range table.R32 {
			for r := rng.Lo; r <= rng.Hi; r += rng.Stride {
				allRunes[rune(r)] = empty
			}
		}
	}
	return allRunes
}

func InternalTest(t *testing.T, csFactory func([]*unicode.RangeTable) source.CharSource) {
	for _, dataset := range Datasets {
		t.Run(dataset.Name, func(t *testing.T) {
			allRunes := tablesToRuneSet(dataset.Tables)
			cs := csFactory(dataset.Tables)
			require.Equal(t, len(allRunes), cs.Count())
			cnt := cs.Count()
			for i := 0; i < cnt; i++ {
				r := cs.At(i)
				_, ok := allRunes[r]
				require.True(t, ok, "rune %d \"%s\" not found", i, string(r))
				delete(allRunes, r)
			}
			require.Equal(t, 0, len(allRunes), allRunes)
		})
	}
}

func InternalBenchmarkAt(b *testing.B, csFactory func([]*unicode.RangeTable) source.CharSource) {
	for _, dataset := range Datasets {
		b.Run(dataset.Name, func(b *testing.B) {
			cs := csFactory(dataset.Tables)
			cnt := cs.Count()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = cs.At(i % cnt)
			}
		})
	}
}

func InternalBenchmarkCreation(b *testing.B, csFactory func([]*unicode.RangeTable) source.CharSource) {
	for _, dataset := range Datasets {
		b.Run(dataset.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = csFactory(dataset.Tables)
			}
		})
	}
}
