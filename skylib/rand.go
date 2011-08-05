package skylib

import (
	"rand"
)

func RandWord(n int) string {
	word := make([]byte, n)
	for i := 0; i < n; i++ {
		r := byte(rand.Intn('z'-'a')) + 'a'
		//fmt.Printf("%d: %c\n", i, r)
		word[i] = r
	}
	return string(word)
}
