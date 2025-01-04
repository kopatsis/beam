package services

import (
	"beam/data/models"
	"beam/data/repositories"
	"fmt"
	"time"
)

type ListService interface {
	GetFavesLine(customerID int, variantID int) (bool, *models.FavesLine, error)
	GetSavesList(customerID int, variantID int) (bool, *models.SavesList, error)

	GetLastOrdersList(customerID int, variantID int) (bool, *models.LastOrdersList, error)
	GetLastOrdersListProd(customerID int, productID int) (bool, *models.LastOrdersList, error)

	AddFavesLine(name string, customerID, variantID int, ps *productService) error
	AddSavesList(name string, customerID, variantID int, ps *productService) error

	DeleteFavesLine(name string, customerID, variantID int, ps *productService) error
	DeleteSavesList(name string, customerID, variantID int, ps *productService) error

	UpdateLastOrdersList(customerID int, orderDate time.Time, orderID string, variants map[int]int) error

	GetFavesLineByPage(name string, customerID, page int, ps *productService) (models.FavesListRender, error)
	GetSavesListByPage(name string, customerID, page int, ps *productService) (models.SavesListRender, error)
	GetLastOrdersListByPage(name string, customerID, page int, ps *productService) (models.LastOrderListRender, error)
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

func (s *listService) GetFavesLine(customerID int, variantID int) (bool, *models.FavesLine, error) {
	panic("not implemented")
}

func (s *listService) GetSavesList(customerID int, variantID int) (bool, *models.SavesList, error) {
	panic("not implemented")
}

func (s *listService) GetLastOrdersList(customerID int, variantID int) (bool, *models.LastOrdersList, error) {
	panic("not implemented")
}

func (s *listService) GetLastOrdersListProd(customerID int, productID int) (bool, *models.LastOrdersList, error) {
	panic("not implemented")
}

func (s *listService) UpdateLastOrdersList(customerID int, orderDate time.Time, orderID string, variants map[int]int) error {
	panic("not implemented")
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
		}
	}

	ret.Data = data
	ret.Prev = prev
	ret.Next = next
	return ret, nil
}
