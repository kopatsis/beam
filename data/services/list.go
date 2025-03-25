package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/listhelp"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"
)

type ListService interface {
	GetFavesLine(dpi *DataPassIn, variantID int, ps ProductService) (bool, error)
	GetSavesList(dpi *DataPassIn, variantID int, ps ProductService) (bool, error)

	CreateCustomList(dpi *DataPassIn, name string) (bool, int, error)
	CreateCustomListWithVar(dpi *DataPassIn, name string, variantID int, ps ProductService) (bool, int, error)
	ChangeCustomListName(dpi *DataPassIn, listID int, name string) error
	ArchiveCustomList(dpi *DataPassIn, listID int) error

	GetLastOrdersList(dpi *DataPassIn, variantID int, ps ProductService) (bool, *models.LastOrdersList, error)
	GetLastOrdersListProd(dpi *DataPassIn, productID int, ps ProductService) (bool, *models.LastOrdersList, error)

	AddFavesLine(dpi *DataPassIn, variantID int, ps ProductService) error
	AddSavesList(dpi *DataPassIn, variantID int, ps ProductService) error
	DeleteFavesLine(dpi *DataPassIn, variantID int, ps ProductService) (string, *models.LimitedVariantRedis, error)
	DeleteSavesList(dpi *DataPassIn, variantID int, ps ProductService) (string, *models.LimitedVariantRedis, error)

	AddFavesLineRender(dpi *DataPassIn, variantID int, page int, ps ProductService) (models.FavesListRender, error)
	AddSavesListRender(dpi *DataPassIn, variantID int, page int, ps ProductService) (models.SavesListRender, error)
	DeleteFavesLineRender(dpi *DataPassIn, variantID int, page int, ps ProductService) (models.FavesListRender, error)
	DeleteSavesListRender(dpi *DataPassIn, variantID int, page int, ps ProductService) (models.SavesListRender, error)

	UpdateLastOrdersList(dpi *DataPassIn, orderDate time.Time, orderID string, vids []int, ps ProductService) error

	GetFavesListByPage(dpi *DataPassIn, page int, ps ProductService) (models.FavesListRender, error)
	GetSavesListByPage(dpi *DataPassIn, page int, ps ProductService) (models.SavesListRender, error)
	GetLastOrdersListByPage(dpi *DataPassIn, page int, ps ProductService) (models.LastOrderListRender, error)
	GetCustomListByPage(dpi *DataPassIn, page, listID int, ps ProductService) (models.CustomListRender, error)

	CartToSavesList(dpi *DataPassIn, lineID int, ps ProductService, cs CartService) (models.SavesListRender, *models.CartRender, error)

	AddToCustomList(dpi *DataPassIn, variantID int, listID int, ps ProductService) error
	DeleteFromCustomList(dpi *DataPassIn, variantID int, listID int, ps ProductService) (string, *models.LimitedVariantRedis, error)
	DeleteFromCustomListAndRender(dpi *DataPassIn, variantID, listID, page int, ps ProductService) (models.CustomListRender, error)

	RetrieveAllCustomLists(dpi *DataPassIn, fromURL url.Values) (models.AllCustomLists, error)
	RetrieveAllCustomListsWithVar(dpi *DataPassIn, fromURL url.Values, variantID int, ps ProductService) (models.AllCustomLists, models.LimitedVariantRedis, error)
	RetrieveCustomListsForVars(dpi *DataPassIn, variantID int, ps ProductService) (models.AllListsForVariant, error)

	RetrieveAllListsAndCounts(dpi *DataPassIn, fromURL url.Values) (models.AllListsAndCounts, error)

	SetCustomPublicStatus(dpi *DataPassIn, listID int, public bool) error
	RenderSharedCustomList(providedCustID string, page, listID int, ps ProductService) (models.CustomListRender, bool, error)
	ShareCustomList(dpi *DataPassIn, listID int) (int, string, error)

	UndoFavesDelete(dpi *DataPassIn, variantID int, dateSt string, page int, ps ProductService, cs CustomerService) (models.FavesListRender, error)
	UndoSavesDelete(dpi *DataPassIn, variantID int, dateSt string, page int, ps ProductService, cs CustomerService) (models.SavesListRender, error)
	UndoCustomDelete(dpi *DataPassIn, listID, variantID int, dateSt string, page int, ps ProductService) (models.CustomListRender, error)
}

type listService struct {
	listRepo repositories.ListRepository
}

func NewListService(listRepo repositories.ListRepository) ListService {
	return &listService{listRepo: listRepo}
}

func (s *listService) AddFavesLine(dpi *DataPassIn, variantID int, ps ProductService) error {
	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
	if err != nil {
		return err
	} else if len(lvs) != 1 {
		return fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	return s.listRepo.AddFavesLine(dpi.CustomerID, lvs[0].ProductID, variantID, false, time.Time{})
}

func (s *listService) AddSavesList(dpi *DataPassIn, variantID int, ps ProductService) error {
	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
	if err != nil {
		return err
	} else if len(lvs) != 1 {
		return fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	return s.listRepo.AddSavesList(dpi.CustomerID, lvs[0].ProductID, variantID, false, time.Time{})
}

func (s *listService) DeleteFavesLine(dpi *DataPassIn, variantID int, ps ProductService) (string, *models.LimitedVariantRedis, error) {
	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
	if err != nil {
		return "", nil, err
	} else if len(lvs) != 1 {
		return "", nil, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	added, alreadyDeleted, err := s.listRepo.DeleteFavesLine(dpi.CustomerID, variantID)
	if err != nil {
		return "", nil, err
	} else if alreadyDeleted {
		return "", nil, nil
	}

	return config.EncodeTime(added), lvs[0], nil
}

func (s *listService) DeleteSavesList(dpi *DataPassIn, variantID int, ps ProductService) (string, *models.LimitedVariantRedis, error) {
	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
	if err != nil {
		return "", nil, err
	} else if len(lvs) != 1 {
		return "", nil, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	added, alreadyDeleted, err := s.listRepo.DeleteSavesList(dpi.CustomerID, variantID)
	if err != nil {
		return "", nil, err
	} else if alreadyDeleted {
		return "", nil, nil
	}

	return config.EncodeTime(added), lvs[0], nil
}

func (s *listService) AddFavesLineRender(dpi *DataPassIn, variantID int, page int, ps ProductService) (models.FavesListRender, error) {
	err := s.AddFavesLine(dpi, variantID, ps)
	if err != nil {
		return models.FavesListRender{}, err
	}
	return s.GetFavesListByPage(dpi, page, ps)
}

func (s *listService) AddSavesListRender(dpi *DataPassIn, variantID int, page int, ps ProductService) (models.SavesListRender, error) {
	err := s.AddSavesList(dpi, variantID, ps)
	if err != nil {
		return models.SavesListRender{}, err
	}
	return s.GetSavesListByPage(dpi, page, ps)
}

func (s *listService) DeleteFavesLineRender(dpi *DataPassIn, variantID int, page int, ps ProductService) (models.FavesListRender, error) {
	useDate, variant, err := s.DeleteFavesLine(dpi, variantID, ps)

	if variant == nil && err != nil {
		return models.FavesListRender{}, err
	} else if err != nil {
		log.Printf("Issue updating time for custom list deletion; error: %v; store: %s; varid: %d; custid: %d\n", err, dpi.Store, variantID, dpi.CustomerID)
	}

	results, err := s.GetFavesListByPage(dpi, page, ps)
	if err != nil {
		return results, err
	}

	if variant != nil {
		results.Deletion = &models.ListDeletionRender{
			Variant: *variant,
			DateSt:  useDate,
		}
		if useDate == "" {
			results.Deletion.DateSt = config.EncodeTime(time.Now())
		}
	}

	return results, nil
}

func (s *listService) DeleteSavesListRender(dpi *DataPassIn, variantID int, page int, ps ProductService) (models.SavesListRender, error) {
	useDate, variant, err := s.DeleteSavesList(dpi, variantID, ps)

	if variant == nil && err != nil {
		return models.SavesListRender{}, err
	} else if err != nil {
		log.Printf("Issue updating time for custom list deletion; error: %v; store: %s; varid: %d; custid: %d\n", err, dpi.Store, variantID, dpi.CustomerID)
	}

	results, err := s.GetSavesListByPage(dpi, page, ps)
	if err != nil {
		return results, err
	}

	if variant != nil {
		results.Deletion = &models.ListDeletionRender{
			Variant: *variant,
			DateSt:  useDate,
		}
		if useDate == "" {
			results.Deletion.DateSt = config.EncodeTime(time.Now())
		}
	}

	return results, nil
}

func (s *listService) GetFavesLine(dpi *DataPassIn, variantID int, ps ProductService) (bool, error) {
	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
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

func (s *listService) GetSavesList(dpi *DataPassIn, variantID int, ps ProductService) (bool, error) {
	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
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

func (s *listService) GetLastOrdersList(dpi *DataPassIn, variantID int, ps ProductService) (bool, *models.LastOrdersList, error) {
	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
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

func (s *listService) GetLastOrdersListProd(dpi *DataPassIn, productID int, ps ProductService) (bool, *models.LastOrdersList, error) {

	in, v, err := s.listRepo.CheckLastOrdersList(dpi.CustomerID, productID)
	if err != nil {
		return false, nil, err
	} else if !in {
		return false, nil, nil
	}

	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{v.ProductID})
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

	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, vids)
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

func (s *listService) GetFavesListByPage(dpi *DataPassIn, page int, ps ProductService) (models.FavesListRender, error) {
	ret := models.FavesListRender{}

	list, prev, next, err := s.listRepo.GetFavesListByPage(dpi.CustomerID, page)
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

	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, vids)
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

	ct, err := s.listRepo.GetFavesListCount(dpi.CustomerID)
	if err != nil {
		return ret, err
	}
	ret.Count = ct

	return ret, nil
}

func (s *listService) GetSavesListByPage(dpi *DataPassIn, page int, ps ProductService) (models.SavesListRender, error) {
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

	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, vids)
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

	ct, err := s.listRepo.GetSavesListCount(dpi.CustomerID)
	if err != nil {
		return ret, err
	}
	ret.Count = ct

	return ret, nil
}

func (s *listService) GetLastOrdersListByPage(dpi *DataPassIn, page int, ps ProductService) (models.LastOrderListRender, error) {
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

	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, vids)
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

	ct, err := s.listRepo.GetLastOrderListCount(dpi.CustomerID)
	if err != nil {
		return ret, err
	}
	ret.Count = ct

	return ret, nil
}

func (s *listService) GetCustomListByPage(dpi *DataPassIn, page, listID int, ps ProductService) (models.CustomListRender, error) {
	ret := models.CustomListRender{}

	list, prev, next, err := s.listRepo.GetCustomListLineByPage(dpi.CustomerID, page, listID)
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

	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, vids)
	if err != nil {
		return ret, err
	} else if len(lvs) != len(list) {
		return ret, fmt.Errorf("incorrect length match for fetching limited variants")
	}

	data := []*models.CustomListLineRender{}
	for _, l := range list {
		data = append(data, &models.CustomListLineRender{
			CustomLine: *l,
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

	ct, err := s.listRepo.GetCustomListCount(dpi.CustomerID, listID)
	if err != nil {
		return ret, err
	}
	ret.Count = ct

	return ret, nil
}

func (s *listService) CartToSavesList(dpi *DataPassIn, lineID int, ps ProductService, cs CartService) (models.SavesListRender, *models.CartRender, error) {

	line, err := cs.GetCartLineWithValidation(dpi, lineID)
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

// Reached maximum, listID, error
func (s *listService) CreateCustomList(dpi *DataPassIn, name string) (bool, int, error) {
	unmaxed, err := s.listRepo.CheckCustomListCount(dpi.CustomerID)
	if err != nil {
		return false, 0, err
	} else if !unmaxed {
		return true, 0, nil
	}

	id, err := s.listRepo.CreateCustomList(dpi.CustomerID, name)
	if err != nil {
		return false, 0, err
	}
	return false, id, nil
}

func (s *listService) CreateCustomListWithVar(dpi *DataPassIn, name string, variantID int, ps ProductService) (bool, int, error) {
	maxReached, listID, err := s.CreateCustomList(dpi, name)
	if maxReached || err != nil {
		return maxReached, listID, err
	}

	if err := s.AddToCustomList(dpi, variantID, listID, ps); err != nil {
		return false, listID, err
	}

	return false, listID, nil
}

func (s *listService) ChangeCustomListName(dpi *DataPassIn, listID int, name string) error {

	if len(name) > 140 {
		name = name[:139]
	}

	return s.listRepo.UpdateCustomListTitle(listID, dpi.CustomerID, name)
}

func (s *listService) ArchiveCustomList(dpi *DataPassIn, listID int) error {
	if err := s.listRepo.ArchiveCustomList(listID, dpi.CustomerID); err != nil {
		return err
	}

	return s.listRepo.SetCustomLastUpdated(dpi.CustomerID, listID)
}

func (s *listService) AddToCustomList(dpi *DataPassIn, variantID int, listID int, ps ProductService) error {
	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
	if err != nil {
		return err
	} else if len(lvs) != 1 {
		return fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	if _, err := s.listRepo.GetSingleCustomList(dpi.CustomerID, listID); err != nil {
		return err
	}

	if err := s.listRepo.AddToCustomList(dpi.CustomerID, listID, lvs[0].VariantID, lvs[0].ProductID, false, time.Time{}); err != nil {
		return err
	}

	return s.listRepo.SetCustomLastUpdated(dpi.CustomerID, listID)
}

func (s *listService) DeleteFromCustomList(dpi *DataPassIn, variantID int, listID int, ps ProductService) (string, *models.LimitedVariantRedis, error) {
	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
	if err != nil {
		return "", nil, err
	} else if len(lvs) != 1 {
		return "", nil, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	if _, err := s.listRepo.GetSingleCustomList(dpi.CustomerID, listID); err != nil {
		return "", nil, err
	}

	added, alreadyDeleted, err := s.listRepo.DeleteFromCustomList(dpi.CustomerID, listID, lvs[0].VariantID)
	if err != nil {
		return "", nil, err
	} else if alreadyDeleted {
		return "", nil, nil
	}

	return config.EncodeTime(added), lvs[0], s.listRepo.SetCustomLastUpdated(dpi.CustomerID, listID)
}

func (s *listService) DeleteFromCustomListAndRender(dpi *DataPassIn, variantID, listID, page int, ps ProductService) (models.CustomListRender, error) {
	useDate, variant, err := s.DeleteFromCustomList(dpi, variantID, listID, ps)

	if variant == nil && err != nil {
		return models.CustomListRender{}, err
	} else if err != nil {
		log.Printf("Issue updating time for custom list deletion; error: %v; store: %s; varid: %d; custid: %d\n", err, dpi.Store, variantID, dpi.CustomerID)
	}

	results, err := s.GetCustomListByPage(dpi, page, listID, ps)
	if err != nil {
		return results, err
	}

	if variant != nil {
		results.Deletion = &models.ListDeletionRender{
			Variant: *variant,
			DateSt:  useDate,
		}
		if useDate == "" {
			results.Deletion.DateSt = config.EncodeTime(time.Now())
		}
	}

	return results, nil
}

func (s *listService) RetrieveAllCustomLists(dpi *DataPassIn, fromURL url.Values) (models.AllCustomLists, error) {
	ret := models.AllCustomLists{Lists: []models.CustomListRenderBrief{}}

	sort, desc := listhelp.ParseQueryParams(fromURL)

	lists, err := s.listRepo.GetCustomListsForCustomer(dpi.CustomerID)
	if err != nil {
		return ret, err
	}

	idList := make([]int, len(lists))
	for i, l := range lists {
		idList[i] = l.ID
		ret.Lists = append(ret.Lists, models.CustomListRenderBrief{
			CustomList: l,
		})
	}

	counts, err := s.listRepo.CountsForCustomLists(dpi.CustomerID, idList)
	if err != nil {
		return ret, err
	}

	for i, lb := range ret.Lists {
		lb.Count = counts[lb.CustomList.ID]
		ret.Lists[i] = lb
	}

	ret.SortBy(sort, desc)

	return ret, nil
}

func (s *listService) RetrieveAllCustomListsWithVar(dpi *DataPassIn, fromURL url.Values, variantID int, ps ProductService) (models.AllCustomLists, models.LimitedVariantRedis, error) {
	list, err := s.RetrieveAllCustomLists(dpi, fromURL)
	if err != nil {
		return list, models.LimitedVariantRedis{}, err
	}

	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
	if err != nil {
		return list, models.LimitedVariantRedis{}, err
	} else if len(lvs) != 1 {
		return list, models.LimitedVariantRedis{}, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	return list, *lvs[0], nil
}

func (s *listService) RetrieveCustomListsForVars(dpi *DataPassIn, variantID int, ps ProductService) (models.AllListsForVariant, error) {
	ret := models.AllListsForVariant{
		VariantID: variantID,
		Customs:   []models.CustomListForVariant{},
	}

	favesIn := false
	favesErr, customErr := error(nil), error(nil)
	var statuses map[int]bool

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		lists, err := s.listRepo.GetCustomListsForCustomer(dpi.CustomerID)
		if err != nil {
			customErr = err
			return
		}

		idList := make([]int, len(lists))
		for i, l := range lists {
			idList[i] = l.ID
			ret.Customs = append(ret.Customs, models.CustomListForVariant{
				CustomList: l,
			})
		}

		statuses, customErr = s.listRepo.HasVariantInLists(dpi.CustomerID, variantID, idList)
	}()

	go func() {
		defer wg.Done()
		favesIn, favesErr = s.GetFavesLine(dpi, variantID, ps)
	}()

	wg.Wait()

	if favesErr != nil && customErr != nil {
		return ret, fmt.Errorf("errors from both faves list and custom lists; faves list: %v; custom lists: %v", favesErr, customErr)
	}

	if customErr == nil {
		ret.FavesHasVar = favesIn
	}

	if customErr == nil {
		for i, lb := range ret.Customs {
			lb.HasVar = statuses[lb.CustomList.ID]
			ret.Customs[i] = lb
		}
		ret.Sort()
		if len(ret.Customs) >= config.MAX_CUSTOM_LISTS {
			ret.CanAddAnother = false
		} else {
			ret.CanAddAnother = true
		}
	}

	if favesErr != nil {
		return ret, favesErr
	}

	return ret, customErr
}

func (s *listService) RetrieveAllListsAndCounts(dpi *DataPassIn, fromURL url.Values) (models.AllListsAndCounts, error) {
	ret := models.AllListsAndCounts{}

	customs, err := s.RetrieveAllCustomLists(dpi, fromURL)
	if err != nil {
		return ret, err
	}
	ret.AllCustomLists = customs

	favesErr, lastErr := error(nil), error(nil)
	var wait sync.WaitGroup
	wait.Add(2)

	go func() {
		defer wait.Done()
		ret.FavesCount, favesErr = s.listRepo.GetFavesListCount(dpi.CustomerID)
	}()

	go func() {
		defer wait.Done()
		ret.LastOrderListCount, lastErr = s.listRepo.GetLastOrderListCount(dpi.CustomerID)
	}()

	if favesErr != nil && lastErr != nil {
		return ret, fmt.Errorf("errors from both faves count and last orders list count; faves: %v; lastorders: %v", favesErr, lastErr)
	} else if favesErr != nil {
		return ret, favesErr
	}
	return ret, lastErr
}

func (s *listService) SetCustomPublicStatus(dpi *DataPassIn, listID int, public bool) error {
	return s.listRepo.SetCustomLastUpdated(dpi.CustomerID, listID)
}

func (s *listService) ShareCustomList(dpi *DataPassIn, listID int) (int, string, error) {
	if err := s.SetCustomPublicStatus(dpi, listID, true); err != nil {
		return 0, "", err
	}

	return listID, config.EncryptInt(dpi.CustomerID), nil
}

// Render, is public, error
func (s *listService) RenderSharedCustomList(providedCustID string, page, listID int, ps ProductService) (models.CustomListRender, bool, error) {

	unencryptedID, err := config.DecryptInt(providedCustID)
	if err != nil {
		return models.CustomListRender{}, false, err
	}

	cl, err := s.GetCustomListByPage(&DataPassIn{CustomerID: unencryptedID}, page, listID, ps)
	if err != nil {
		return models.CustomListRender{}, false, err
	}

	if !cl.CustomList.Public {
		return models.CustomListRender{}, false, nil
	}

	return cl, true, nil
}

func (s *listService) UndoFavesDelete(dpi *DataPassIn, variantID int, dateSt string, page int, ps ProductService, cs CustomerService) (models.FavesListRender, error) {
	realDate, err := config.DecodeTime(dateSt)
	if err != nil {
		return models.FavesListRender{}, err
	}

	cust, err := cs.GetCustomerByID(dpi.CustomerID)
	if err != nil {
		return models.FavesListRender{}, err
	}

	if realDate.Before(cust.Created) {
		realDate = cust.Created
	}

	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
	if err != nil {
		return models.FavesListRender{}, err
	} else if len(lvs) != 1 {
		return models.FavesListRender{}, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	if err := s.listRepo.AddFavesLine(dpi.CustomerID, lvs[0].ProductID, variantID, true, realDate); err != nil {
		return models.FavesListRender{}, err
	}

	return s.GetFavesListByPage(dpi, page, ps)
}

func (s *listService) UndoSavesDelete(dpi *DataPassIn, variantID int, dateSt string, page int, ps ProductService, cs CustomerService) (models.SavesListRender, error) {
	realDate, err := config.DecodeTime(dateSt)
	if err != nil {
		return models.SavesListRender{}, err
	}

	cust, err := cs.GetCustomerByID(dpi.CustomerID)
	if err != nil {
		return models.SavesListRender{}, err
	}

	if realDate.Before(cust.Created) {
		realDate = cust.Created
	}

	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
	if err != nil {
		return models.SavesListRender{}, err
	} else if len(lvs) != 1 {
		return models.SavesListRender{}, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	if err := s.listRepo.AddSavesList(dpi.CustomerID, lvs[0].ProductID, variantID, true, realDate); err != nil {
		return models.SavesListRender{}, err
	}

	return s.GetSavesListByPage(dpi, page, ps)
}

func (s *listService) UndoCustomDelete(dpi *DataPassIn, listID, variantID int, dateSt string, page int, ps ProductService) (models.CustomListRender, error) {
	realDate, err := config.DecodeTime(dateSt)
	if err != nil {
		return models.CustomListRender{}, err
	}

	list, err := s.listRepo.GetSingleCustomList(dpi.CustomerID, listID)
	if err != nil {
		return models.CustomListRender{}, err
	}

	if realDate.Before(list.Created) {
		realDate = list.Created
	}

	lvs, err := ps.GetLimitedVariants(dpi, dpi.Store, []int{variantID})
	if err != nil {
		return models.CustomListRender{}, err
	} else if len(lvs) != 1 {
		return models.CustomListRender{}, fmt.Errorf("could not find single lim var for id: %d", variantID)
	}

	if err := s.listRepo.AddToCustomList(dpi.CustomerID, listID, lvs[0].VariantID, lvs[0].ProductID, true, realDate); err != nil {
		return models.CustomListRender{}, err
	}

	return s.GetCustomListByPage(dpi, page, listID, ps)
}
