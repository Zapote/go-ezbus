package ezbus

import (
	"errors"
	"fmt"
)

func retry(fn func(), attempts int) (err error) {
	defer func() {
		if r := recover(); r != nil && attempts > 1 {
			err = errors.New(fmt.Sprintf("%v", r))
			retry(fn, attempts-1)
		} else {
			attempts = 0
		}
	}()

	fn()

	return err
}
