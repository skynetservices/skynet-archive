package skylib

import "testing"

func testRandWordN(t *testing.T, n int) {
	word := RandWord(n)
	if n != len(word) {
		t.Errorf("Expected %d, got %d\n", n, len(word))
	}
}
func TestRandWord(t *testing.T) {
	for _, i := range []int{0, 1, 2, 99} {
		testRandWordN(t, i)
	}
}
