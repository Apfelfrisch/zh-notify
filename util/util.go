package util

import (
	"encoding/json"
	"fmt"
)

func Debug(value any) {
	formatted, err := json.MarshalIndent(value, "", "  ")

	if err != nil {
		panic(err)
	}

	fmt.Println(string(formatted))
}
