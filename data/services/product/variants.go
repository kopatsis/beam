package product

import (
	"beam/data/models"
)

func NameVariant(product models.ProductRedis, vid int) string {
	name := product.Title
	for _, v := range product.Variants {
		if v.PK == vid {
			if v.Var1Value != "" && v.Var1Value != "*" {
				name += " / " + v.Var1Value
			}
			if v.Var2Value != "" && v.Var2Value != "*" {
				name += " / " + v.Var2Value
			}
			if v.Var3Value != "" && v.Var3Value != "*" {
				name += " / " + v.Var3Value
			}
		}
	}
	return name
}

func VariantSelectorRenders(product models.ProductRedis, vid int) models.AllVariants {
	layer, val1, val2, val3, inv := createBaseVars(product, vid)

	first, second, third := createInitialLists(product, vid, inv, val1, val2, val3)

	first, second, third = modifyEachList(product, vid, inv, layer, val1, val2, val3, first, second, third)

	ret := models.AllVariants{
		First:     first,
		FirstKey:  product.Var1Key,
		Second:    second,
		SecondKey: product.Var2Key,
		Third:     third,
		ThirdKey:  product.Var3Key,
	}
	return ret
}

func checkIfInBlock(list []models.VariantBlock, val string) bool {
	for _, b := range list {
		if b.Name == val {
			return true
		}
	}
	return false
}

func createBaseVars(product models.ProductRedis, vid int) (int, string, string, string, int) {
	layer := 0
	val1, val2, val3 := "", "", ""
	inv := -1

	for _, v := range product.Variants {
		if v.PK == vid {
			inv = v.Quantity
			val1 = v.Var1Value
			val2 = v.Var2Value
			val3 = v.Var3Value
			if val1 != "" {
				layer++
			}
			if val2 != "" {
				layer++
			}
			if val3 != "" {
				layer++
			}
		}
	}

	return layer, val1, val2, val3, inv
}

func createInitialLists(product models.ProductRedis, vid, inv int, val1, val2, val3 string) ([]models.VariantBlock, []models.VariantBlock, []models.VariantBlock) {
	first, second, third := []models.VariantBlock{}, []models.VariantBlock{}, []models.VariantBlock{}

	for _, v := range product.Variants {
		if v.Var1Value != "" && v.Var1Value != "*" && !checkIfInBlock(first, v.Var1Value) {
			first = append(first, models.VariantBlock{
				Name:     v.Var1Value,
				Selected: v.Var1Value == val1,
				Stocked:  v.Var1Value == val1 && inv > 0,
			})
			if first[len(first)-1].Stocked {
				first[len(first)-1].VariantID = vid
			}
		}
		if v.Var2Value != "" && v.Var1Value == val1 && !checkIfInBlock(second, v.Var2Value) {
			second = append(second, models.VariantBlock{
				Name:     v.Var2Value,
				Selected: v.Var2Value == val2,
				Stocked:  v.Var2Value == val2 && inv > 0,
			})
			if second[len(second)-1].Stocked {
				second[len(second)-1].VariantID = vid
			}
		}
		if v.Var3Value != "" && v.Var1Value == val1 && v.Var2Value == val2 && !checkIfInBlock(third, v.Var3Value) {
			third = append(third, models.VariantBlock{
				Name:     v.Var3Value,
				Selected: v.Var3Value == val3,
				Stocked:  v.Var3Value == val3 && inv > 0,
			})
			if third[len(third)-1].Stocked {
				third[len(third)-1].VariantID = vid
			}
		}
	}

	return first, second, third
}

func modifyEachList(product models.ProductRedis, vid, inv, layer int, val1, val2, val3 string, first, second, third []models.VariantBlock) ([]models.VariantBlock, []models.VariantBlock, []models.VariantBlock) {

	if layer == 1 {

		for i, b := range first {
			if !b.Selected {
				for _, v := range product.Variants {
					if v.Var1Value == b.Name {
						b.VariantID = v.PK
						b.Stocked = v.Quantity > 0
						break
					}
				}
				first[i] = b
			}
		}

	} else if layer == 2 {

		for i, b := range second {
			if !b.Selected {
				for _, v := range product.Variants {
					if v.Var2Value == b.Name && v.Var1Value == val1 {
						b.VariantID = v.PK
						b.Stocked = v.Quantity > 0
						break
					}
				}
				second[i] = b
			}
		}

		for i, b := range first {
			if !b.Selected {
				completed := false
				for _, v := range product.Variants {
					if v.Var1Value == b.Name && v.Var2Value == val2 && v.Quantity > 0 {
						b.VariantID = v.PK
						b.Stocked = v.Quantity > 0
						completed = true
						break
					}
				}

				if !completed {
					for _, v := range product.Variants {
						if v.Var1Value == b.Name && v.Quantity > 0 {
							b.VariantID = v.PK
							b.Stocked = v.Quantity > 0
							completed = true
							break
						}
					}
				}

				if !completed {
					for _, v := range product.Variants {
						if v.Var1Value == b.Name && v.Var2Value == val2 {
							b.VariantID = v.PK
							b.Stocked = v.Quantity > 0
							completed = true
							break
						}
					}
				}

				if !completed {
					for _, v := range product.Variants {
						if v.Var1Value == b.Name {
							b.VariantID = v.PK
							b.Stocked = v.Quantity > 0
							completed = true
							break
						}
					}
				}

				first[i] = b
			}
		}
	} else {
		for i, b := range third {
			if !b.Selected {
				for _, v := range product.Variants {
					if v.Var3Value == b.Name && v.Var1Value == val1 && v.Var2Value == val2 {
						b.VariantID = v.PK
						b.Stocked = v.Quantity > 0
						break
					}
				}
				third[i] = b
			}
		}

		for i, b := range second {
			if !b.Selected {
				completed := false
				for _, v := range product.Variants {
					if v.Var2Value == b.Name && v.Var3Value == val3 && v.Var1Value == val1 && v.Quantity > 0 {
						b.VariantID = v.PK
						b.Stocked = v.Quantity > 0
						completed = true
						break
					}
				}

				if !completed {
					for _, v := range product.Variants {
						if v.Var1Value == b.Name && v.Var1Value == val1 && v.Quantity > 0 {
							b.VariantID = v.PK
							b.Stocked = v.Quantity > 0
							completed = true
							break
						}
					}
				}

				if !completed {
					for _, v := range product.Variants {
						if v.Var2Value == b.Name && v.Var3Value == val3 && v.Var1Value == val1 {
							b.VariantID = v.PK
							b.Stocked = v.Quantity > 0
							completed = true
							break
						}
					}
				}

				if !completed {
					for _, v := range product.Variants {
						if v.Var1Value == b.Name && v.Var1Value == val1 {
							b.VariantID = v.PK
							b.Stocked = v.Quantity > 0
							completed = true
							break
						}
					}
				}

				second[i] = b
			}
		}

		for i, b := range first {
			if !b.Selected {
				completed := false
				for _, v := range product.Variants {
					if v.Var1Value == b.Name && v.Var3Value == val3 && v.Var2Value == val2 && v.Quantity > 0 {
						b.VariantID = v.PK
						b.Stocked = v.Quantity > 0
						completed = true
						break
					}
				}

				if !completed {
					for _, v := range product.Variants {
						if v.Var1Value == b.Name && v.Var2Value == val2 && v.Quantity > 0 {
							b.VariantID = v.PK
							b.Stocked = v.Quantity > 0
							completed = true
							break
						}
					}
				}

				if !completed {
					for _, v := range product.Variants {
						if v.Var1Value == b.Name && v.Quantity > 0 {
							b.VariantID = v.PK
							b.Stocked = v.Quantity > 0
							completed = true
							break
						}
					}
				}

				if !completed {
					for _, v := range product.Variants {
						if v.Var1Value == b.Name && v.Var3Value == val3 && v.Var2Value == val2 {
							b.VariantID = v.PK
							b.Stocked = v.Quantity > 0
							completed = true
							break
						}
					}
				}

				if !completed {
					for _, v := range product.Variants {
						if v.Var1Value == b.Name && v.Var2Value == val2 {
							b.VariantID = v.PK
							b.Stocked = v.Quantity > 0
							completed = true
							break
						}
					}
				}

				if !completed {
					for _, v := range product.Variants {
						if v.Var1Value == b.Name {
							b.VariantID = v.PK
							b.Stocked = v.Quantity > 0
							completed = true
							break
						}
					}
				}

				first[i] = b
			}
		}
	}

	return first, second, third
}
