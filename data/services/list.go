package services

import (
	"beam/data/models"
	"beam/data/repositories"
	"fmt"
	"time"
)

type ListService interface {
	GetFavesLine(dpi *DataPassIn, variantID int, ps *productService) (bool, error)
	GetSavesList(dpi *DataPassIn, variantID int, ps *productService) (bool, error)

	GetLastOrdersList(dpi *DataPassIn, variantID int, ps *productService) (bool, *models.LastOrdersList, error)
	GetLastOrdersListProd(dpi *DataPassIn, productID int, ps *productService) (bool, *models.LastOrdersList, error)

	AddFavesLine(dpi *DataPassIn, variantID int, ps *productService) error
	AddSavesList(dpi *DataPassIn, variantID int, ps *productService) error
	DeleteFavesLine(dpi *DataPassIn, variantID int, ps *productService) error
	DeleteSavesList(dpi *DataPassIn, variantID int, ps *productService) error

	AddFavesLineRender(dpi *DataPassIn, variantID int, page int, ps *productService) (models.FavesListRender, error)
	AddSavesListRender(dpi *DataPassIn, variantID int, page int, ps *productService) (models.SavesListRender, error)
	DeleteFavesLineRender(dpi *DataPassIn, variantID int, page int, ps *productService) (models.FavesListRender, error)
	DeleteSavesListRender(dpi *DataPassIn, variantID int, page int, ps *productService) (models.SavesListRender, error)

	UpdateLastOrdersList(dpi *DataPassIn, orderDate time.Time, orderID string, vids []int, ps ProductService) error

	GetFavesLineByPage(dpi *DataPassIn, page int, ps *productService) (models.FavesListRender, error)
	GetSavesListByPage(dpi *DataPassIn, page int, ps *productService) (models.SavesListRender, error)
	GetLastOrdersListByPage(dpi *DataPassIn, page int, ps *productService) (models.LastOrderListRender, error)

	CartToSavesList(dpi *DataPassIn, lineID int, ps *productService, cs *cartService) (models.SavesListRender, *models.CartRender, error)
}

type listService struct {
	listRepo repositories.ListRepository
}

func NewListService(listRepo repositories.ListRepository) ListService {
	return &listService{listRepo: listRepo}
}

func (s *listService) AddFavesLine(dpi *DataPassIn, variantID int, ps *productService) error {
	lvs, err := ps.GetLimitedVariants(dpi.Store, []int{variantID})
	if err != nil {
		return err
	} else if len(lvs) != 1 {
		return fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	return s.listRepo.AddFavesLine(dpi.CustomerID, lvs[0].ProductID, variantID)
}

func (s *listService) AddSavesList(dpi *DataPassIn, variantID int, ps *productService) error {
	lvs, err := ps.GetLimitedVariants(dpi.Store, []int{variantID})
	if err != nil {
		return err
	} else if len(lvs) != 1 {
		return fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	return s.listRepo.AddSavesList(dpi.CustomerID, lvs[0].ProductID, variantID)
}

func (s *listService) DeleteFavesLine(dpi *DataPassIn, variantID int, ps *productService) error {
	lvs, err := ps.GetLimitedVariants(dpi.Store, []int{variantID})
	if err != nil {
		return err
	} else if len(lvs) != 1 {
		return fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	return s.listRepo.DeleteFavesLine(dpi.CustomerID, variantID)
}

func (s *listService) DeleteSavesList(dpi *DataPassIn, variantID int, ps *productService) error {
	lvs, err := ps.GetLimitedVariants(dpi.Store, []int{variantID})
	if err != nil {
		return err
	} else if len(lvs) != 1 {
		return fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	return s.listRepo.DeleteSavesList(dpi.CustomerID, variantID)
}

func (s *listService) AddFavesLineRender(dpi *DataPassIn, variantID int, page int, ps *productService) (models.FavesListRender, error) {
	err := s.AddFavesLine(dpi, variantID, ps)
	if err != nil {
		return models.FavesListRender{}, err
	}
	return s.GetFavesLineByPage(dpi, page, ps)
}

func (s *listService) AddSavesListRender(dpi *DataPassIn, variantID int, page int, ps *productService) (models.SavesListRender, error) {
	err := s.AddSavesList(dpi, variantID, ps)
	if err != nil {
		return models.SavesListRender{}, err
	}
	return s.GetSavesListByPage(dpi, page, ps)
}

func (s *listService) DeleteFavesLineRender(dpi *DataPassIn, variantID int, page int, ps *productService) (models.FavesListRender, error) {
	err := s.DeleteFavesLine(dpi, variantID, ps)
	if err != nil {
		return models.FavesListRender{}, err
	}
	return s.GetFavesLineByPage(dpi, page, ps)
}

func (s *listService) DeleteSavesListRender(dpi *DataPassIn, variantID int, page int, ps *productService) (models.SavesListRender, error) {
	err := s.DeleteSavesList(dpi, variantID, ps)
	if err != nil {
		return models.SavesListRender{}, err
	}
	return s.GetSavesListByPage(dpi, page, ps)
}

func (s *listService) GetFavesLine(dpi *DataPassIn, variantID int, ps *productService) (bool, error) {
	lvs, err := ps.GetLimitedVariants(dpi.Store, []int{variantID})
	if err != nil {
		return false, err
	} else if len(lvs) != 1 {
		return false, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	in, _, err := s.listRepo.CheckFavesLine(dpi.CustomerID, variantID)
	if err != nil {
		return false, err
	}

	return in, nil
}

func (s *listService) GetSavesList(dpi *DataPassIn, variantID int, ps *productService) (bool, error) {
	lvs, err := ps.GetLimitedVariants(dpi.Store, []int{variantID})
	if err != nil {
		return false, err
	} else if len(lvs) != 1 {
		return false, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	in, _, err := s.listRepo.CheckSavesList(dpi.CustomerID, variantID)
	if err != nil {
		return false, err
	}

	return in, nil
}

func (s *listService) GetLastOrdersList(dpi *DataPassIn, variantID int, ps *productService) (bool, *models.LastOrdersList, error) {
	lvs, err := ps.GetLimitedVariants(dpi.Store, []int{variantID})
	if err != nil {
		return false, nil, err
	} else if len(lvs) != 1 {
		return false, nil, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	in, v, err := s.listRepo.CheckLastOrdersList(dpi.CustomerID, variantID)
	if err != nil {
		return false, nil, err
	} else if !in {
		return false, nil, nil
	}

	return true, v, nil
}

func (s *listService) GetLastOrdersListProd(dpi *DataPassIn, productID int, ps *productService) (bool, *models.LastOrdersList, error) {

	in, v, err := s.listRepo.CheckLastOrdersList(dpi.CustomerID, productID)
	if err != nil {
		return false, nil, err
	} else if !in {
		return false, nil, nil
	}

	lvs, err := ps.GetLimitedVariants(dpi.Store, []int{v.ProductID})
	if err != nil {
		return false, nil, err
	} else if len(lvs) != 1 {
		return false, nil, fmt.Errorf("could not find single lim var for id: %d", v.ProductID)
	}

	return true, v, nil
}

func (s *listService) UpdateLastOrdersList(dpi *DataPassIn, orderDate time.Time, orderID string, vids []int, ps ProductService) error {
	if len(vids) == 0 {
		return nil
	}

	lvs, err := ps.GetLimitedVariants(dpi.Store, vids)
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

	return s.listRepo.UpdateLastOrdersList(dpi.CustomerID, orderDate, orderID, use)
}

func (s *listService) GetFavesLineByPage(dpi *DataPassIn, page int, ps *productService) (models.FavesListRender, error) {
	ret := models.FavesListRender{}

	list, prev, next, err := s.listRepo.GetFavesLineByPage(dpi.CustomerID, page)
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

	lvs, err := ps.GetLimitedVariants(dpi.Store, vids)
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

func (s *listService) GetSavesListByPage(dpi *DataPassIn, page int, ps *productService) (models.SavesListRender, error) {
	ret := models.SavesListRender{}

	list, prev, next, err := s.listRepo.GetSavesListByPage(dpi.CustomerID, page)
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

	lvs, err := ps.GetLimitedVariants(dpi.Store, vids)
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

func (s *listService) GetLastOrdersListByPage(dpi *DataPassIn, page int, ps *productService) (models.LastOrderListRender, error) {
	ret := models.LastOrderListRender{}

	list, prev, next, err := s.listRepo.GetLastOrdersListByPage(dpi.CustomerID, page)
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

	lvs, err := ps.GetLimitedVariants(dpi.Store, vids)
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

func (s *listService) CartToSavesList(dpi *DataPassIn, lineID int, ps *productService, cs *cartService) (models.SavesListRender, *models.CartRender, error) {

	line, err := cs.cartRepo.GetCartLineWithValidation(dpi.CustomerID, dpi.CartID, lineID)
	if err != nil {
		return models.SavesListRender{}, nil, err
	}

	cr, err := cs.AdjustQuantity(dpi, lineID, 0, ps)
	if err != nil {
		return models.SavesListRender{}, nil, err
	}

	sl, err := s.AddSavesListRender(dpi, line.VariantID, 1, ps)
	if err != nil {
		return models.SavesListRender{}, nil, err
	}

	return sl, cr, nil
}
