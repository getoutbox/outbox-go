package outbox

import "fmt"

func errEmpty(op string) error {
	return fmt.Errorf("outbox: %s: server returned empty response", op)
}
