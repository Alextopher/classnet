package classnet

import "math/rand"

// Returns a random key, value pair from a map
func RandomEntry[K comparable, V any](m map[K]V) (K, V) {
	n := rand.Intn(len(m))
	var k K
	var v V
	for k, v = range m {
		if n == 0 {
			break
		}
		n--
	}
	return k, v
}
