package core

import "fmt"

// Checks if panic is due to closed channel
func IsClosedChanPanic(r any) bool {
	if r == nil {
		return false
	}
	return fmt.Sprint(r) == "send on closed channel"
}
