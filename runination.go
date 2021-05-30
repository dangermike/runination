package runination

import (
	"math/rand"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dangermike/runination/flatsource"
	"github.com/dangermike/runination/rangedsource"
	"github.com/dangermike/runination/source"
)

type RandomGenerator struct {
	charSource      source.CharSource
	estBytesPerChar float64
}

func NewRanged(tables []*unicode.RangeTable) *RandomGenerator {
	return New(rangedsource.New(tables))
}

func NewFlat(tables []*unicode.RangeTable) *RandomGenerator {
	return New(flatsource.New(tables))
}

func New(charSource source.CharSource) *RandomGenerator {
	totalBytes := 0
	totalChars := 0
	if charSource.Count() < 1000 {
		totalChars += charSource.Count()
		for i := 0; i < charSource.Count(); i++ {
			totalBytes += utf8.RuneLen(charSource.At(i))
		}
	} else {
		// we don't want to read thousands of chars, so we'll sample instead
		totalChars = 500
		for i := 0; i < totalChars; i++ {
			totalBytes += utf8.RuneLen(charSource.At(rand.Intn(charSource.Count())))
		}
	}
	return &RandomGenerator{
		charSource:      charSource,
		estBytesPerChar: float64(totalBytes) / float64(totalChars),
	}
}

// String generates a string of chars from the allowed set of length between
// min and max, inclusive
func (rg *RandomGenerator) String(min, max int) string {
	strlen := min
	if min < max {
		strlen += rand.Intn(1 + max - min)
	}
	sb := strings.Builder{}
	if targetSize := int(rg.estBytesPerChar * float64(strlen)); sb.Cap() < targetSize {
		sb.Grow(targetSize)
	}

	cnt := rg.charSource.Count()
	totalBytes := 0
	for i := 0; i < strlen; i++ {
		n, _ := sb.WriteRune(rg.charSource.At(rand.Intn(cnt)))
		totalBytes += n
	}
	return sb.String()
}

func (rg *RandomGenerator) EstBytesPerChar() float64 {
	return rg.estBytesPerChar
}
