package system

import (
	"fmt"
	"github.com/suiyunonghen/DxCommonLib/system/goid"
	"testing"
)

func TestGetRoutineId(t *testing.T) {
	fmt.Println(GetRoutineId())

	fmt.Println(goid.GetGoID())
}