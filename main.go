package main

import (
	"fmt"
	"github.com/lithammer/fuzzysearch/fuzzy"
	_ "github.com/lithammer/fuzzysearch/fuzzy"
)

func main() {
	//bootstrap.Main()
	fuzzy.Match("twl", "cartwheel")  // true
	fuzzy.Match("cart", "cartwheel") // true
	fuzzy.Match("cw", "cartwheel")   // true
	fuzzy.Match("ee", "cartwheel")   // true
	fuzzy.Match("art", "cartwheel")  // true
	fuzzy.Match("eeel", "cartwheel") // false
	fuzzy.Match("dog", "cartwheel")  // false
	fuzzy.Match("kitten", "sitting") // false

	fuzzy.RankMatch("kitten", "sitting") // -1
	fuzzy.RankMatch("cart", "cartwheel") // 5

	words := []string{"cartwheel", "foobar", "wheel", "baz"}
	fuzzy.Find("whl", words) // [cartwheel wheel]

	fuzzy.RankFind("whl", words) // [{whl cartwheel 6 0} {whl wheel 2 2}]

	// Unicode normalized matching.
	fuzzy.MatchNormalized("cartwheel", "cartwhéél") // true

	// Case insensitive matching.
	fuzzy.MatchFold("ArTeeL", "cartwheel") // true
	fmt.Println(fuzzy.RankFind("whl", words))
}
