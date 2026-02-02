package parse

import "strconv"

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
