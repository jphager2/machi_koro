package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func JoinInts(ints []int, delimiter string) string {
	str := make([]string, len(ints))
	for i, v := range ints {
		str[i] = strconv.Itoa(v)
	}

	return strings.Join(str, delimiter)
}

func GetInt(validValues []int) (int, error) {
	var val int
	fmt.Scan(&val)

	for _, v := range validValues {
		if v == val {
			return val, nil
		}
	}

	return 0, errors.New(fmt.Sprintf("Invalid input '%d' for values (%s)", val, JoinInts(validValues, ",")))
}
