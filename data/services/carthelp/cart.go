package carthelp

import "beam/data/models"

func UpdateCartQuant(cart models.CartRender) int {
	ret := 0
	for _, l := range cart.CartLines {
		ret += l.ActualLine.Quantity
	}
	return ret
}
