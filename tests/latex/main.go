// +build gofuzz

package fuzz

import "github.com/feedmepos/canvas"

// Fuzz is a fuzz test.
func Fuzz(data []byte) int {
	_, _ = canvas.ParseLaTeX(string(data))
	return 1
}
