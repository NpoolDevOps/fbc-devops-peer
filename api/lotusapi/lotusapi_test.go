package lotusapi

import (
	"fmt"
	"testing"
)

func TestLotusapi(t *testing.T) {
	height, err := ChainHeadHeight("10.155.8.32")
	if err != nil {
		fmt.Println("err is", err)
		return
	}
	fmt.Println("height is", height)
}
