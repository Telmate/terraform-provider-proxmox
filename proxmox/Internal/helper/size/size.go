package size

import (
	"regexp"
	"strconv"
)

var Regex = regexp.MustCompile(`^[123456789]\d*[KMGT]?$`)

const (
	kibiByte int64 = 1
	mebiByte int64 = 1024
	gibiByte int64 = 1048576
	tebiByte int64 = 1073741824
)

func String(kibibytes int64) string {
	if kibibytes%tebiByte == 0 {
		return strconv.FormatInt(kibibytes/tebiByte, 10) + "T"
	}
	if kibibytes%gibiByte == 0 {
		return strconv.FormatInt(kibibytes/gibiByte, 10) + "G"
	}
	if kibibytes%mebiByte == 0 {
		return strconv.FormatInt(kibibytes/mebiByte, 10) + "M"
	}
	return strconv.FormatInt(kibibytes, 10) + "K"
}

// Relies on the input being validated
func Parse_Unsafe(size string) int {
	if len(size) > 1 {
		switch size[len(size)-1:] {
		case "T":
			return parseSize_Unsafe(size, tebiByte)
		case "G":
			return parseSize_Unsafe(size, gibiByte)
		case "M":
			return parseSize_Unsafe(size, mebiByte)
		case "K":
			return parseSize_Unsafe(size, kibiByte)
		}
	}
	tmpSize, _ := strconv.ParseInt(size, 10, 0)
	return int(tmpSize * gibiByte)
}

// Relies on the input being validated
func parseSize_Unsafe(size string, multiplier int64) int {
	tmpSize, _ := strconv.ParseInt(size[:len(size)-1], 10, 0)
	return int(tmpSize * multiplier)
}
