package carthelp

import "beam/data/models"

func UpdateCartSub(cart *models.CartRender) {
	quant, subtotal := 0, 0
	for i, l := range cart.CartLines {
		l.ActualLine.Subtotal = l.ActualLine.Price * l.ActualLine.Quantity
		subtotal += l.ActualLine.Subtotal
		quant += l.ActualLine.Quantity
		cart.CartLines[i].ActualLine = l.ActualLine
	}
	cart.SumQuantity = quant
	cart.Subtotal = subtotal
}
