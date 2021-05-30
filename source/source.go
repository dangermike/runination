package source

type CharSource interface {
	Count() int
	At(int) rune
}
