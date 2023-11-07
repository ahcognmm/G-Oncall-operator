package customerror

import "fmt"

// HTTPError wraps a raw HTTP error
type HTTPError struct {
	StatusCode int
	HtErr      error
}

func (htErr HTTPError) Error() string {
	return fmt.Sprintf("HTTP error %d: %s.", htErr.StatusCode, htErr.HtErr)
}
