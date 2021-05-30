# Runination (roon-in-ation)

A random string generator that works with unicode ranges

## What is it

There is a lot of content on StackOverflow and the like explaining how to quickly generate random strings. These almost all start with a string to hold the set of characters available. This works well for cases where single-byte utf-8 strings (read: ASCII-7) are used. When multi-byte characters appear, it turns to a big mess. Not only that, the set of characters is pretty staggering -- the number of characters in the [GraphicRanges](https://golang.org/pkg/unicode/#IsGraphic) range is ~146,000. Having that as a constant in a file seems kind of silly.

Runination, as you might guess from the name, is [rune](https://golang.org/pkg/builtin/#rune)-based. By taking in a slice of [RangeTable](https://golang.org/pkg/unicode/#RangeTable)s, it is not necessary to define all of the characters in a constant a the start.

This whole package is ridiculous overkill and was build purely for fun. Going this deep on a problem like this is really not necessary.

## CharSource

A [CharSource](source/source.go) is what translates the [RangeTable](https://golang.org/pkg/unicode/#RangeTable)s into runes. A RangeTable defines how to walk the set of valid codepoints, which you can imagine doing two ways: One could walk them as soon as you get them, storing all of the runes encountered in a slice. Alternatively, you could navigate to the nth codepoint on-demand. The first method is faster once constructed, but can use a lot of memory. The second method has a faster start-up time, but uses no additional memory. These two methods are implemented as [CharSourceFlat](source/flatsource/flat.go) and [CharSourceRanged](source/rangedsource/ranged.go).

As expected, the flat source has a longer creation time in all but the smallest of cases and uses a lot more memory:

```plain
$ go test ./... -bench Benchmark.+Creation -benchmem
goos: darwin
goarch: amd64
pkg: github.com/dangermike/runination/flatsource
cpu: Intel(R) Core(TM) i7-1068NG7 CPU @ 2.30GHz
BenchmarkFlatCreation/ASCII_hex_digit-8                 14417836                84.09 ns/op          120 B/op          2 allocs/op
BenchmarkFlatCreation/ASCII_Alphanumeric-8               9743510               122.3 ns/op           280 B/op          2 allocs/op
BenchmarkFlatCreation/Latin-8                             684128              1622 ns/op            6168 B/op          2 allocs/op
BenchmarkFlatCreation/Symbol-8                            133599              8781 ns/op           32792 B/op          2 allocs/op
BenchmarkFlatCreation/GraphicRanges-8                       7039            147511 ns/op          581657 B/op          2 allocs/op

PASS
ok      github.com/dangermike/runination/flatsource     6.241s
goos: darwin
goarch: amd64
pkg: github.com/dangermike/runination/rangedsource
cpu: Intel(R) Core(TM) i7-1068NG7 CPU @ 2.30GHz
BenchmarkRangedCreation/ASCII_hex_digit-8                9320533               127.0 ns/op           184 B/op          4 allocs/op
BenchmarkRangedCreation/ASCII_Alphanumeric-8             9511452               126.9 ns/op           184 B/op          4 allocs/op
BenchmarkRangedCreation/Latin-8                          4308471               277.9 ns/op           768 B/op          4 allocs/op
BenchmarkRangedCreation/Symbol-8                          843031              1388 ns/op            5360 B/op          5 allocs/op
BenchmarkRangedCreation/GraphicRanges-8                   148989              7750 ns/op           36208 B/op          5 allocs/op
PASS
ok      github.com/dangermike/runination/rangedsource   6.732s
```

However, it also a lot faster. Flat is O(1), whereas ranged is O(log2 n) + O(m) where n is the number of ranges and m is the average size of each range.

```plain
$ go test ./... -bench Benchmark.+At -benchmem

goos: darwin
goarch: amd64
pkg: github.com/dangermike/runination/flatsource
cpu: Intel(R) Core(TM) i7-1068NG7 CPU @ 2.30GHz
BenchmarkFlatAt/ASCII_hex_digit-8               412119396                2.851 ns/op           0 B/op          0 allocs/op
BenchmarkFlatAt/ASCII_Alphanumeric-8            441814352                2.634 ns/op           0 B/op          0 allocs/op
BenchmarkFlatAt/Latin-8                         451034258                2.603 ns/op           0 B/op          0 allocs/op
BenchmarkFlatAt/Symbol-8                        445885848                2.632 ns/op           0 B/op          0 allocs/op
BenchmarkFlatAt/GraphicRanges-8                 454417116                2.635 ns/op           0 B/op          0 allocs/op
PASS
ok      github.com/dangermike/runination/flatsource     7.464s
goos: darwin
goarch: amd64
pkg: github.com/dangermike/runination/rangedsource
cpu: Intel(R) Core(TM) i7-1068NG7 CPU @ 2.30GHz
BenchmarkRangedAt/ASCII_hex_digit-8             100000000               11.50 ns/op            0 B/op          0 allocs/op
BenchmarkRangedAt/ASCII_Alphanumeric-8          93092020                11.45 ns/op            0 B/op          0 allocs/op
BenchmarkRangedAt/Latin-8                       62280600                17.45 ns/op            0 B/op          0 allocs/op
BenchmarkRangedAt/Symbol-8                      47110302                24.17 ns/op            0 B/op          0 allocs/op
BenchmarkRangedAt/GraphicRanges-8               37516694                29.71 ns/op            0 B/op          0 allocs/op
PASS
ok      github.com/dangermike/runination/rangedsource   5.847s
```

Even just in terms of time, the choice isn't clear. As an example, if you are using the whole GraphicRanges set, you have the following functions:

```plain
flat = 2.635n + 147511
ranged = 29.71n + 7750
```

Which means the break-even point is at 5,162 characters, This is ignoring the fact that Flat uses 16x more memory than Ranged. On the other hand, for the `ASCII_hex_digit` set, Flat is faster and uses less memory before you even start!

## Getting random strings

Now that we have a source for random characters, we need to put them into strings. Most of the tricks one might use for fixed-byte characters (e.g. masking) just aren't available in variable-width characters. The only choice available is [strings.Builder](https://golang.org/pkg/strings/#Builder). There is one last trick we can use: the Builder, like slice and lots of other vector-like structures, grows when it runs out of space. If we can give it a good guess on the target size, we can avoid extra allocations. With fixed-width characters, that's as easy as multiplying the number of characters by the byte width. With variable-width characters, a guess is the best we can do. The `RandomGenerator` inspects the `CharSource` on start-up, measuring characters from the target set. If the total set of characters is less than 1,000, it just reads them all (less than 20µs on my laptop). If it is larger, 500 characters are randomly sampled. The result, which you can see in the benchmarks below, is that we average only one allocation per cycle. Without this optimization, only calling `Builder.Grow` with the length of the string, we averaged 3-4 allocations when using the largest strings and the GraphicRanges character set.

```plain
$ go test ./... -bench BenchmarkRandomString -benchmem

goos: darwin
goarch: amd64
pkg: github.com/dangermike/runination
cpu: Intel(R) Core(TM) i7-1068NG7 CPU @ 2.30GHz
BenchmarkRandomString/ASCII_hex_digit/flat_10-10-8               5531746               216.9 ns/op            16 B/op          1 allocs/op
BenchmarkRandomString/ASCII_hex_digit/ranged_10-10-8             3276037               364.8 ns/op            16 B/op          1 allocs/op
BenchmarkRandomString/ASCII_hex_digit/flat_10-15-8               4058647               286.0 ns/op            16 B/op          1 allocs/op
BenchmarkRandomString/ASCII_hex_digit/ranged_10-15-8             2506100               466.5 ns/op            16 B/op          1 allocs/op
BenchmarkRandomString/ASCII_hex_digit/flat_50-100-8               772429              1568 ns/op              82 B/op          1 allocs/op
BenchmarkRandomString/ASCII_hex_digit/ranged_50-100-8             472190              2494 ns/op              82 B/op          1 allocs/op
BenchmarkRandomString/ASCII_hex_digit/flat_512-512-8              115868             10263 ns/op             512 B/op          1 allocs/op
BenchmarkRandomString/ASCII_hex_digit/ranged_512-512-8             72187             16754 ns/op             512 B/op          1 allocs/op
BenchmarkRandomString/ASCII_Alphanumeric/flat_10-10-8            5443030               218.2 ns/op            16 B/op          1 allocs/op
BenchmarkRandomString/ASCII_Alphanumeric/ranged_10-10-8          3251095               370.9 ns/op            16 B/op          1 allocs/op
BenchmarkRandomString/ASCII_Alphanumeric/flat_10-15-8            4063056               302.4 ns/op            16 B/op          1 allocs/op
BenchmarkRandomString/ASCII_Alphanumeric/ranged_10-15-8          2507072               480.8 ns/op            16 B/op          1 allocs/op
BenchmarkRandomString/ASCII_Alphanumeric/flat_50-100-8            771399              1558 ns/op              82 B/op          1 allocs/op
BenchmarkRandomString/ASCII_Alphanumeric/ranged_50-100-8          427035              2566 ns/op              82 B/op          1 allocs/op
BenchmarkRandomString/ASCII_Alphanumeric/flat_512-512-8           116931             10252 ns/op             512 B/op          1 allocs/op
BenchmarkRandomString/ASCII_Alphanumeric/ranged_512-512-8          70140             17178 ns/op             512 B/op          1 allocs/op
BenchmarkRandomString/Latin/flat_10-10-8                         3841479               308.5 ns/op            70 B/op          1 allocs/op
BenchmarkRandomString/Latin/ranged_10-10-8                       1996341               592.1 ns/op            82 B/op          1 allocs/op
BenchmarkRandomString/Latin/flat_10-15-8                         2983227               396.3 ns/op            96 B/op          1 allocs/op
BenchmarkRandomString/Latin/ranged_10-15-8                       1626229               738.1 ns/op            93 B/op          1 allocs/op
BenchmarkRandomString/Latin/flat_50-100-8                         587740              1974 ns/op             428 B/op          1 allocs/op
BenchmarkRandomString/Latin/ranged_50-100-8                       296396              3996 ns/op             443 B/op          1 allocs/op
BenchmarkRandomString/Latin/flat_512-512-8                         94171             12717 ns/op            2207 B/op          1 allocs/op
BenchmarkRandomString/Latin/ranged_512-512-8                       45918             26341 ns/op            1649 B/op          1 allocs/op
BenchmarkRandomString/Symbol/flat_10-10-8                        3467898               342.2 ns/op           104 B/op          1 allocs/op
BenchmarkRandomString/Symbol/ranged_10-10-8                      1522629               776.3 ns/op           104 B/op          1 allocs/op
BenchmarkRandomString/Symbol/flat_10-15-8                        2964027               403.5 ns/op           110 B/op          1 allocs/op
BenchmarkRandomString/Symbol/ranged_10-15-8                      1265169               956.3 ns/op           116 B/op          1 allocs/op
BenchmarkRandomString/Symbol/flat_50-100-8                        567274              2054 ns/op             413 B/op          1 allocs/op
BenchmarkRandomString/Symbol/ranged_50-100-8                      224954              5318 ns/op             572 B/op          1 allocs/op
BenchmarkRandomString/Symbol/flat_512-512-8                        88570             13291 ns/op            2342 B/op          1 allocs/op
BenchmarkRandomString/Symbol/ranged_512-512-8                      32974             36527 ns/op            5119 B/op          1 allocs/op
BenchmarkRandomString/GraphicRanges/flat_10-10-8                 3347757               359.8 ns/op           106 B/op          1 allocs/op
BenchmarkRandomString/GraphicRanges/ranged_10-10-8               1537748               781.0 ns/op            87 B/op          1 allocs/op
BenchmarkRandomString/GraphicRanges/flat_10-15-8                 2726622               440.4 ns/op           116 B/op          1 allocs/op
BenchmarkRandomString/GraphicRanges/ranged_10-15-8               1217469               979.5 ns/op           119 B/op          1 allocs/op
BenchmarkRandomString/GraphicRanges/flat_50-100-8                 499150              2294 ns/op             606 B/op          1 allocs/op
BenchmarkRandomString/GraphicRanges/ranged_50-100-8               216877              5493 ns/op             524 B/op          1 allocs/op
BenchmarkRandomString/GraphicRanges/flat_512-512-8                 84614             14123 ns/op            2070 B/op          1 allocs/op
BenchmarkRandomString/GraphicRanges/ranged_512-512-8               31282             37012 ns/op            5859 B/op          1 allocs/op
PASS
ok      github.com/dangermike/runination        60.034s
```

## Credits

This was a silly weekend project, but I was inspired to use the unicode RangeTables by [chrismcguire/gobberish](https://github.com/chrismcguire/gobberish). Runination doesn't add any functionality beyond what's in `gobberish`, but it is a lot faster. I don't know why one might care, but it is.
