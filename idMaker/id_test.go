package idMaker

import (
	"fmt"
	"testing"
)

func TestGenerateId(t *testing.T) {
	for i := 0; i < 5000; i++ {
		id := GenerateId()
		fmt.Printf("%d %b %d \n", i, id, id)
	}
}
