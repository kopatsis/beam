package product

import "math"

func VolumeDiscPrice(price, quantity int, allowed bool) int {
	if !allowed || quantity < 5 {
		return price
	} else if quantity < 10 {
		return int(math.Round(float64(price) * 0.95))
	} else if quantity < 20 {
		return int(math.Round(float64(price) * 0.9))
	} else if quantity < 50 {
		return int(math.Round(float64(price) * 0.85))
	} else {
		return int(math.Round(float64(price) * 0.8))
	}
}
