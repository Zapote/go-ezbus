package ezbus

import (
	"fmt"
	"log"
)

func retry(fn func() error, attempts int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Recovered from panic: %v", r)
		}
	}()

	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		log.Printf("Attempt failed: %s", err.Error())
	}

	return err
}
