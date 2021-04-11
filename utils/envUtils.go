package utils

import (
	"fmt"
	"os"
)

// Get an environment variable or panic if it is not set.
func GetEnvOrPanic(key string) string {
	val, present := os.LookupEnv(key)
	if !present {
		panic(fmt.Sprintf("\"%s\" environment variable is not present", key))
	}

	return val
}
