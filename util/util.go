package util

import (
	"encoding/json"
	"fmt"
)

func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func Debug(value any) {
	formatted, err := json.MarshalIndent(value, "", "  ")

	if err != nil {
		panic(err)
	}

	fmt.Println(string(formatted))
}
