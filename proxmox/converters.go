package proxmox

import (
	"strconv"
)

const (
	kibibyte int64 = 1
	mebibyte int64 = 1024
	gibibyte int64 = 1048576
	tebibyte int64 = 1073741824
)

func convert_KibibytesToString(kibibytes int64) string {
	if kibibytes%tebibyte == 0 {
		return strconv.FormatInt(kibibytes/tebibyte, 10) + "T"
	}
	if kibibytes%gibibyte == 0 {
		return strconv.FormatInt(kibibytes/gibibyte, 10) + "G"
	}
	if kibibytes%mebibyte == 0 {
		return strconv.FormatInt(kibibytes/mebibyte, 10) + "M"
	}
	return strconv.FormatInt(kibibytes, 10) + "K"
}

// Relies on the input being validated
func convert_SizeStringToKibibytes_Unsafe(size string) int {
	if len(size) > 1 {
		switch size[len(size)-1:] {
		case "T":
			return parseSize_Unsafe(size, tebibyte)
		case "G":
			return parseSize_Unsafe(size, gibibyte)
		case "M":
			return parseSize_Unsafe(size, mebibyte)
		case "K":
			return parseSize_Unsafe(size, kibibyte)
		}
	}
	tmpSize, _ := strconv.ParseInt(size, 10, 0)
	return int(tmpSize * gibibyte)
}

// Relies on the input being validated
func parseSize_Unsafe(size string, multiplier int64) int {
	tmpSize, _ := strconv.ParseInt(size[:len(size)-1], 10, 0)
	return int(tmpSize * multiplier)
}
