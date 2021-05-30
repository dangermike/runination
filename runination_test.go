package runination_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/dangermike/runination"
	"github.com/dangermike/runination/testhelper"
	"github.com/stretchr/testify/require"
)

func TestEstBytesPerChar(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for _, test := range []struct {
		name   string
		tables []*unicode.RangeTable
		estBpC float64
		fudge  float64
	}{
		{"asciihex", []*unicode.RangeTable{unicode.ASCII_Hex_Digit}, 1.0, 0.0},
		{"graphic", unicode.GraphicRanges, 3.6, 0.25},
	} {
		t.Run(test.name, func(t *testing.T) {
			actBpC := runination.NewRanged(test.tables).EstBytesPerChar()
			require.LessOrEqual(t, test.estBpC-test.fudge, actBpC)
			require.LessOrEqual(t, actBpC, test.estBpC+test.fudge)
		})
	}
}

func TestSimple(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	for _, dataset := range testhelper.Datasets {
		t.Run(dataset.Name, func(t *testing.T) {
			for _, test := range []struct {
				name            string
				randomGenerator *runination.RandomGenerator
				min             int
				max             int
			}{
				{"flat", runination.NewFlat(dataset.Tables), 10, 10},
				{"ranged", runination.NewRanged(dataset.Tables), 10, 10},
				{"flat", runination.NewFlat(dataset.Tables), 10, 15},
				{"ranged", runination.NewRanged(dataset.Tables), 10, 15},
				{"flat", runination.NewFlat(dataset.Tables), 50, 100},
				{"ranged", runination.NewRanged(dataset.Tables), 50, 100},
			} {
				lengths := map[int]int{}
				last := ""
				for i := 0; i < 1000; i++ {
					s := test.randomGenerator.NewString(test.min, test.max)
					require.NotEqual(t, last, s)
					lengths[utf8.RuneCountInString(s)] += 1
					last = s
				}
				for i := test.min; i <= test.max; i++ {
					require.NotEqual(t, 0, lengths[i], "length: %d", i)
				}
				for k := range lengths {
					require.LessOrEqual(t, test.min, k)
					require.LessOrEqual(t, k, test.max)
				}
			}
		})
	}
}

func BenchmarkRandomString(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	for _, dataset := range testhelper.Datasets {
		b.Run(dataset.Name, func(b *testing.B) {
			for _, test := range []struct {
				name string
				gen  func(int, int) string
				min  int
				max  int
			}{
				{"flat", runination.NewFlat(dataset.Tables).NewString, 10, 10},
				{"ranged", runination.NewRanged(dataset.Tables).NewString, 10, 10},
				{"flat", runination.NewFlat(dataset.Tables).NewString, 10, 15},
				{"ranged", runination.NewRanged(dataset.Tables).NewString, 10, 15},
				{"flat", runination.NewFlat(dataset.Tables).NewString, 50, 100},
				{"ranged", runination.NewRanged(dataset.Tables).NewString, 50, 100},
				{"flat", runination.NewFlat(dataset.Tables).NewString, 512, 512},
				{"ranged", runination.NewRanged(dataset.Tables).NewString, 512, 512},
			} {
				b.Run(test.name+fmt.Sprintf(" %d-%d", test.min, test.max), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						_ = test.gen(test.min, test.max)
					}
				})
			}
		})
	}
}
