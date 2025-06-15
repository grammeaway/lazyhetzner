package utility 

import (
	"strconv"
)

func GetNumberForIndex(index int) string {
	if index == 9 { // 10th item (0-indexed 9)
		return "0"
	}
	return strconv.Itoa(index + 1)
}

func GetIndexFromNumber(keyStr string) int {
	switch keyStr {
	case "1":
		return 0
	case "2":
		return 1
	case "3":
		return 2
	case "4":
		return 3
	case "5":
		return 4
	case "6":
		return 5
	case "7":
		return 6
	case "8":
		return 7
	case "9":
		return 8
	case "0":
		return 9
	default:
		return -1
	}
}


// Helper function for min (Go 1.21+)
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
