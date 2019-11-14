package ezbus

import (
	"fmt"
	"log"
)

func receive(fn func() error, attempts int) (err error) {
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
				log.Printf("Attempt failed: %s", err.Error())
				continue
			}
		}
		return nil
	}
	return err
}
