package parse

import (
	"strconv"
	"strings"
)

func ID(id, prefix string) (int, error) {
	var num int
	var err error
	if len(id) > len(prefix) && id[0:len(prefix)] == prefix {
		num, err = strconv.Atoi(id[len(prefix):])
	} else {
		num, err = strconv.Atoi(id)
	}
	return num, err
}

func DoubleDelimitedString[T any](in string, convert func(string) T) []T {
	out := make([]T, 0)
	if in == "" {
		return out
	}
	stringIter := strings.SplitSeq(in, ",")
	for e := range stringIter {
		tagSubArrays := strings.Split(e, ";")
		if len(tagSubArrays) > 1 {
			tmpOut := make([]T, len(tagSubArrays))
			for i := range tagSubArrays {
				tmpOut[i] = convert(tagSubArrays[i])
			}
			out = append(out, tmpOut...)
		} else {
			out = append(out, convert(e))
		}
	}
	return out
}
