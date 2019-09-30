package ezbus

import "log"

func retry(fn func() error, attempts int) error {
	err := fn()
	attempts--
	if attempts == 0 {
		return err
	}

	if err != nil {
		log.Printf("Attempt failed: %s", err.Error())
		return retry(fn, attempts)
	}

	return nil
}
