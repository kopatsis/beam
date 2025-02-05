package services

import (
	"beam/data/models"
	"beam/data/repositories"
	"fmt"
	"time"
)

type ListService interface {
	GetFavesLine(name string, customerID int, variantID int, ps *productService) (bool, error)
	GetSavesList(name string, customerID int, variantID int, ps *productService) (bool, error)

	GetLastOrdersList(name string, customerID int, variantID int, ps *productService) (bool, *models.LastOrdersList, error)
	GetLastOrdersListProd(name string, customerID int, productID int, ps *productService) (bool, *models.LastOrdersList, error)

	AddFavesLine(name string, customerID, variantID int, ps *productService) error
	AddSavesList(name string, customerID, variantID int, ps *productService) error
	DeleteFavesLine(name string, customerID, variantID int, ps *productService) error
	DeleteSavesList(name string, customerID, variantID int, ps *productService) error

	AddFavesLineRender(name string, customerID, variantID int, page int, ps *productService) (models.FavesListRender, error)
	AddSavesListRender(name string, customerID, variantID int, page int, ps *productService) (models.SavesListRender, error)
	DeleteFavesLineRender(name string, customerID, variantID int, page int, ps *productService) (models.FavesListRender, error)
	DeleteSavesListRender(name string, customerID, variantID int, page int, ps *productService) (models.SavesListRender, error)

	UpdateLastOrdersList(name string, customerID int, orderDate time.Time, orderID string, vids []int, ps *productService) error

	GetFavesLineByPage(name string, customerID, page int, ps *productService) (models.FavesListRender, error)
	GetSavesListByPage(name string, customerID, page int, ps *productService) (models.SavesListRender, error)
	GetLastOrdersListByPage(name string, customerID, page int, ps *productService) (models.LastOrderListRender, error)

	CartToSavesList(dpi DataPassIn, lineID int, ps *productService, cs *cartService) (models.SavesListRender, *models.CartRender, error)
}

type listService struct {
	listRepo repositories.ListRepository
}

func NewListService(listRepo repositories.ListRepository) ListService {
	return &listService{listRepo: listRepo}
}

func (s *listService) AddFavesLine(name string, customerID, variantID int, ps *productService) error {
	lvs, err := ps.GetLimitedVariants(name, []int{variantID})
	if err != nil {
		return err
	} else if len(lvs) != 1 {
		return fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	return s.listRepo.AddFavesLine(customerID, lvs[0].ProductID, variantID)
}

func (s *listService) AddSavesList(name string, customerID, variantID int, ps *productService) error {
	lvs, err := ps.GetLimitedVariants(name, []int{variantID})
	if err != nil {
		return err
	} else if len(lvs) != 1 {
		return fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	return s.listRepo.AddSavesList(customerID, lvs[0].ProductID, variantID)
}

func (s *listService) DeleteFavesLine(name string, customerID, variantID int, ps *productService) error {
	lvs, err := ps.GetLimitedVariants(name, []int{variantID})
	if err != nil {
		return err
	} else if len(lvs) != 1 {
		return fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	return s.listRepo.DeleteFavesLine(customerID, variantID)
}

func (s *listService) DeleteSavesList(name string, customerID, variantID int, ps *productService) error {
	lvs, err := ps.GetLimitedVariants(name, []int{variantID})
	if err != nil {
		return err
	} else if len(lvs) != 1 {
		return fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	return s.listRepo.DeleteSavesList(customerID, variantID)
}

func (s *listService) AddFavesLineRender(name string, customerID, variantID int, page int, ps *productService) (models.FavesListRender, error) {
	err := s.AddFavesLine(name, customerID, variantID, ps)
	if err != nil {
		return models.FavesListRender{}, err
	}
	return s.GetFavesLineByPage(name, customerID, page, ps)
}

func (s *listService) AddSavesListRender(name string, customerID, variantID int, page int, ps *productService) (models.SavesListRender, error) {
	err := s.AddSavesList(name, customerID, variantID, ps)
	if err != nil {
		return models.SavesListRender{}, err
	}
	return s.GetSavesListByPage(name, customerID, page, ps)
}

func (s *listService) DeleteFavesLineRender(name string, customerID, variantID int, page int, ps *productService) (models.FavesListRender, error) {
	err := s.DeleteFavesLine(name, customerID, variantID, ps)
	if err != nil {
		return models.FavesListRender{}, err
	}
	return s.GetFavesLineByPage(name, customerID, page, ps)
}

func (s *listService) DeleteSavesListRender(name string, customerID, variantID int, page int, ps *productService) (models.SavesListRender, error) {
	err := s.DeleteSavesList(name, customerID, variantID, ps)
	if err != nil {
		return models.SavesListRender{}, err
	}
	return s.GetSavesListByPage(name, customerID, page, ps)
}

func (s *listService) GetFavesLine(name string, customerID int, variantID int, ps *productService) (bool, error) {
	lvs, err := ps.GetLimitedVariants(name, []int{variantID})
	if err != nil {
		return false, err
	} else if len(lvs) != 1 {
		return false, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	in, _, err := s.listRepo.CheckFavesLine(customerID, variantID)
	if err != nil {
		return false, err
	}

	return in, nil
}

func (s *listService) GetSavesList(name string, customerID int, variantID int, ps *productService) (bool, error) {
	lvs, err := ps.GetLimitedVariants(name, []int{variantID})
	if err != nil {
		return false, err
	} else if len(lvs) != 1 {
		return false, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	in, _, err := s.listRepo.CheckSavesList(customerID, variantID)
	if err != nil {
		return false, err
	}

	return in, nil
}

func (s *listService) GetLastOrdersList(name string, customerID int, variantID int, ps *productService) (bool, *models.LastOrdersList, error) {
	lvs, err := ps.GetLimitedVariants(name, []int{variantID})
	if err != nil {
		return false, nil, err
	} else if len(lvs) != 1 {
		return false, nil, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	in, v, err := s.listRepo.CheckLastOrdersList(customerID, variantID)
	if err != nil {
		return false, nil, err
	} else if !in {
		return false, nil, nil
	}

	return true, v, nil
}

func (s *listService) GetLastOrdersListProd(name string, customerID int, productID int, ps *productService) (bool, *models.LastOrdersList, error) {

	in, v, err := s.listRepo.CheckLastOrdersList(customerID, productID)
	if err != nil {
		return false, nil, err
	} else if !in {
		return false, nil, nil
	}

	lvs, err := ps.GetLimitedVariants(name, []int{v.ProductID})
	if err != nil {
		return false, nil, err
	} else if len(lvs) != 1 {
		return false, nil, fmt.Errorf("could not find single lim var for id: %d", v.ProductID)
	}

	return true, v, nil
}

func (s *listService) UpdateLastOrdersList(name string, customerID int, orderDate time.Time, orderID string, vids []int, ps *productService) error {
	if len(vids) == 0 {
		return nil
	}

	lvs, err := ps.GetLimitedVariants(name, vids)
	if err != nil {
		return err
	} else if len(lvs) != len(vids) {
		return fmt.Errorf("incorrect length match for fetching limited variants")
	}

	use := map[int]int{}
	for _, vid := range vids {
		found := false
		for _, v := range lvs {
			if v.VariantID == vid {
				use[vid] = v.ProductID
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("could not locate specific variant in limited variants: %d", vid)
		}
	}

	return s.listRepo.UpdateLastOrdersList(customerID, orderDate, orderID, use)
}

func (s *listService) GetFavesLineByPage(name string, customerID, page int, ps *productService) (models.FavesListRender, error) {
	ret := models.FavesListRender{}

	list, prev, next, err := s.listRepo.GetFavesLineByPage(customerID, page)
	if err != nil {
		return ret, err
	}

	vids := []int{}
	for _, l := range list {
		vids = append(vids, l.VariantID)
	}

	if len(vids) == 0 {
		ret.NoData = true
		return ret, nil
	}

	lvs, err := ps.GetLimitedVariants(name, vids)
	if err != nil {
		return ret, err
	} else if len(lvs) != len(list) {
		return ret, fmt.Errorf("incorrect length match for fetching limited variants")
	}

	data := []*models.FavesLineRender{}
	for _, l := range list {
		data = append(data, &models.FavesLineRender{
			FavesLine: *l,
		})
		for _, v := range lvs {
			data[len(data)-1].Variant = *v
			data[len(data)-1].Found = true
			break
		}
	}

	ret.Data = data
	ret.Prev = prev
	ret.Next = next
	return ret, nil
}

func (s *listService) GetSavesListByPage(name string, customerID, page int, ps *productService) (models.SavesListRender, error) {
	ret := models.SavesListRender{}

	list, prev, next, err := s.listRepo.GetSavesListByPage(customerID, page)
	if err != nil {
		return ret, err
	}

	vids := []int{}
	for _, l := range list {
		vids = append(vids, l.VariantID)
	}

	if len(vids) == 0 {
		ret.NoData = true
		return ret, nil
	}

	lvs, err := ps.GetLimitedVariants(name, vids)
	if err != nil {
		return ret, err
	} else if len(lvs) != len(list) {
		return ret, fmt.Errorf("incorrect length match for fetching limited variants")
	}

	data := []*models.SavesLineRender{}
	for _, l := range list {
		data = append(data, &models.SavesLineRender{
			SavesLine: *l,
		})
		for _, v := range lvs {
			data[len(data)-1].Variant = *v
			data[len(data)-1].Found = true
			break
		}
	}

	ret.Data = data
	ret.Prev = prev
	ret.Next = next
	return ret, nil
}

func (s *listService) GetLastOrdersListByPage(name string, customerID, page int, ps *productService) (models.LastOrderListRender, error) {
	ret := models.LastOrderListRender{}

	list, prev, next, err := s.listRepo.GetLastOrdersListByPage(customerID, page)
	if err != nil {
		return ret, err
	}

	vids := []int{}
	for _, l := range list {
		vids = append(vids, l.VariantID)
	}

	if len(vids) == 0 {
		ret.NoData = true
		return ret, nil
	}

	lvs, err := ps.GetLimitedVariants(name, vids)
	if err != nil {
		return ret, err
	} else if len(lvs) != len(list) {
		return ret, fmt.Errorf("incorrect length match for fetching limited variants")
	}

	data := []*models.LastOrderLineRender{}
	for _, l := range list {
		data = append(data, &models.LastOrderLineRender{
			LOLine: *l,
		})
		for _, v := range lvs {
			data[len(data)-1].Variant = *v
			data[len(data)-1].Found = true
			break
		}
	}

	ret.Data = data
	ret.Prev = prev
	ret.Next = next
	return ret, nil
}

func (s *listService) CartToSavesList(dpi DataPassIn, lineID int, ps *productService, cs *cartService) (models.SavesListRender, *models.CartRender, error) {

	line, err := cs.cartRepo.GetCartLineWithValidation(dpi.CustomerID, dpi.CartID, lineID)
	if err != nil {
		return models.SavesListRender{}, nil, err
	}

	cr, err := cs.AdjustQuantity(dpi, lineID, 0, ps)
	if err != nil {
		return models.SavesListRender{}, nil, err
	}

	sl, err := s.AddSavesListRender(dpi.Store, dpi.CustomerID, line.VariantID, 1, ps)
	if err != nil {
		return models.SavesListRender{}, nil, err
	}

	return sl, cr, nil
}
