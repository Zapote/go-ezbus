package ezbus

import (
	"fmt"

	"github.com/zapote/go-ezbus/logger"
)

func receive(messageName string, fn func() error, attempts int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Recovered from panic: %v", r)
		}
	}()

	for i := 0; i < attempts; i++ {
		if err = fn(); err != nil {
			switch v := err.(type) {
			case HandlerNotFoundErr:
				return v
			default:
				logger.Errorf("Attempt #%d, message '%s' failed: %s", (i + 1), messageName, err.Error())
				continue
			}
		}
		return nil
	}
	return err
}
