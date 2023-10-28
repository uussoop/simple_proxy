package utils

import "os"

func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		if fallback == "" {
			panic("no fallback value provided")
		}
		return fallback
	}
	return value
}
