package discount

import (
	"strings"

	"golang.org/x/exp/rand"
)

const Vals = "0123456789abcdefghijklmnopqrstuvwxyz"

func GenerateCartID() string {
	slice := make([]int, 14)

	sum1, sum2 := 0, 0
	for i := range slice {
		v := rand.Intn(36)
		slice[i] = rand.Intn(36)
		if i%2 == 0 {
			sum1 += v * 1
			sum2 += v * 2
		} else {
			sum1 += v * 3
			sum2 += v * 4
		}
	}

	cd1 := sum1 % 36
	cd2 := (sum2 + cd1*2) % 36

	slice = append(slice, cd1)
	slice = append(slice, cd2)

	var builder strings.Builder
	for i, num := range slice {
		if i > 0 && i%4 == 0 {
			builder.WriteString("-")
		}
		builder.WriteByte(Vals[num])
	}
	return builder.String()
}

func CheckID(id string) bool {
	replID := strings.ReplaceAll(id, "-", "")
	if len(replID) != 16 {
		return false
	}

	slice := make([]int, 16)
	for i := range replID {
		index := strings.Index(Vals, string(replID[i]))
		if index == -1 {
			return false
		}
		slice[i] = index
	}

	sum1, sum2 := 0, 0
	for i, v := range slice {
		if i == 14 || i == 15 {
			break
		}
		if i%2 == 0 {
			sum1 += v * 1
			sum2 += v * 2
		} else {
			sum1 += v * 3
			sum2 += v * 4
		}
	}

	if slice[14] != sum1%36 {
		return false
	}

	return slice[15] == (sum2+slice[14]*2)%36
}