package ezbus

import "log"

func retry(fn func() error, attempts int) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		log.Printf("Attempt failed: %s", err.Error())
	}
	return err
}
