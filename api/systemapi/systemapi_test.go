package systemapi

import (
	"testing"
	"fmt"
)

func TestGetTemperature(t *testing.T) {
	tem, err := GetNvmeTemperature("dev")
	fmt.Println("t is", tem, "err is", err)
}
