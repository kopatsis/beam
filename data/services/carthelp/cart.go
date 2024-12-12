package carthelp

import "beam/data/models"

func UpdateCartSub(cart *models.CartRender) {
	quant, subtotal := 0, 0
	for i, l := range cart.CartLines {
		l.Subtotal = l.ActualLine.Price * l.ActualLine.Quantity
		subtotal += l.Subtotal
		quant += l.ActualLine.Quantity
		cart.CartLines[i] = l
	}
	cart.SumQuantity = quant
	cart.Subtotal = subtotal
}
