package services

import (
	"beam/background/emails"
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/product"
	"errors"
	"fmt"
	"log"
	"net/url"
	"slices"
	"sort"
	"strconv"

	"math/rand"
)

type ProductService interface {
	AddProduct(product models.Product) error
	GetProductByID(id int) (*models.Product, error)
	UpdateProduct(product models.Product) error
	DeleteProduct(id int) error

	GetAllProductInfo(fromURL url.Values, Mutex *config.AllMutexes, name string) (models.CollectionRender, error)
	GetProductAndProductRender(Mutex *config.AllMutexes, name, handle, id string) (models.ProductRedis, models.ProductRender, string, error)
	GetProductRender(Mutex *config.AllMutexes, name, handle, id string) (models.ProductRender, string, error)
	GetLimitedVariants(name string, vids []int) ([]*models.LimitedVariantRedis, error)
	GetProductByVariantID(name string, vid int) (models.ProductRedis, string, error)
	GetProductsByVariantIDs(name string, vids []int) (map[int]*models.ProductRedis, error)
	GetProductsMapFromCartLine(name string, cartLines []models.CartLine) (map[int]*models.ProductRedis, error)

	UpdateRatings(dpi *DataPassIn, pid, newRate, oldRate, plusMinus int, tools *config.Tools)
	ConfirmDraftOrderProducts(dpi *DataPassIn, vids []int) (map[int]bool, bool, error)
	RenderComparables(name string, id int) ([]models.ComparablesRender, error)

	SetInventoryFromOrder(dpi *DataPassIn, decrement map[int]int, handles []string) error
}

type productService struct {
	productRepo repositories.ProductRepository
}

func NewProductService(productRepo repositories.ProductRepository) ProductService {
	return &productService{productRepo: productRepo}
}

func (s *productService) AddProduct(product models.Product) error {
	return s.productRepo.Create(product)
}

func (s *productService) GetProductByID(id int) (*models.Product, error) {
	return s.productRepo.Read(id)
}

func (s *productService) UpdateProduct(product models.Product) error {
	return s.productRepo.Update(product)
}

func (s *productService) DeleteProduct(id int) error {
	return s.productRepo.Delete(id)
}

func (s *productService) GetAllProductInfo(fromURL url.Values, Mutex *config.AllMutexes, name string) (models.CollectionRender, error) {
	ret := models.CollectionRender{}

	var query, page, sort string
	otherParams := map[string][]string{}

	for key, values := range fromURL {
		switch key {
		case "qy":
			if len(values) > 0 {
				query = values[0]
			}
		case "pg":
			if len(values) > 0 {
				page = values[0]
			}
		case "st":
			if len(values) > 0 {
				sort = values[0]
			}
		default:
			otherParams[key] = values
		}
	}
	if len(query) > 128 {
		query = query[0:127]
	}

	products, err := s.productRepo.GetAllProductInfo(name)
	if err != nil {
		return ret, err
	}

	endParams := map[string][]string{}
	forTop := models.AllFilters{}
	if len(otherParams) > 0 {
		products, endParams, forTop = product.FilterByTags(otherParams, products, Mutex, name)
	}

	realsort := ""
	if query != "" {
		products, err = product.FuzzySearch(query, products)
		if err != nil {
			return ret, err
		}
	} else {
		realsort, products = product.SortProducts(sort, products)
	}

	var left, pg, right int
	products, left, pg, right = product.PageProducts(page, products)
	fmt.Print(left, right, products, forTop)

	baseURL := product.CreateBasisURL(query, realsort, pg, endParams)

	filterBar, err := product.CreateFilterBar(Mutex, baseURL, name, endParams)
	if err != nil {
		return ret, err
	}

	return models.CollectionRender{
		Products: products,
		URL:      baseURL,
		SideBar:  filterBar,
		TopWords: product.CreateTopWords(forTop, query),
		Paging:   product.PageRender(pg, left, right, baseURL),
	}, nil

}

func (s *productService) GetProductAndProductRender(Mutex *config.AllMutexes, name, handle, id string) (models.ProductRedis, models.ProductRender, string, error) {

	rprod, redir, err := s.productRepo.GetFullProduct(name, handle)
	if err != nil {
		return models.ProductRedis{}, models.ProductRender{}, redir, err
	} else if redir != "" {
		return models.ProductRedis{}, models.ProductRender{}, redir, err
	}
	fmt.Print(rprod, redir)

	actualID := 0
	convertedID, err := strconv.Atoi(id)
	if err == nil {
		for _, v := range rprod.Variants {
			if v.PK == convertedID {
				actualID = v.PK
			}
		}
	}

	if len(rprod.Variants) == 1 && rprod.Var1Key == "&" {
		return rprod, models.ProductRender{
			FullName:  rprod.Title,
			VariantID: rprod.Variants[0].PK,
			Inventory: rprod.Variants[0].Quantity,
			Price:     rprod.Variants[0].Price,
			CompareAt: rprod.Variants[0].CompareAtPrice,
			VarImage:  rprod.Variants[0].VariantImageURL,
		}, "", nil
	}

	if actualID == 0 {
		for _, v := range rprod.Variants {
			if v.Quantity > 0 {
				actualID = v.PK
			}
		}
		if actualID == 0 {
			actualID = rprod.Variants[0].PK
		}
	}

	ret := models.ProductRender{
		FullName:    product.NameVariant(rprod, actualID),
		VariantID:   rprod.Variants[0].PK,
		Inventory:   rprod.Variants[0].Quantity,
		Price:       rprod.Variants[0].Price,
		CompareAt:   rprod.Variants[0].CompareAtPrice,
		VarImage:    rprod.Variants[0].VariantImageURL,
		HasVariants: true,
		Blocks:      product.VariantSelectorRenders(rprod, actualID),
	}

	return rprod, ret, "", nil
}

func (s *productService) GetProductRender(Mutex *config.AllMutexes, name, handle, id string) (models.ProductRender, string, error) {
	_, rend, redir, err := s.GetProductAndProductRender(Mutex, name, handle, id)
	return rend, redir, err
}

func (s *productService) GetLimitedVariants(name string, vids []int) ([]*models.LimitedVariantRedis, error) {
	return s.productRepo.GetLimVars(name, vids)
}

func (s *productService) GetProductByVariantID(name string, vid int) (models.ProductRedis, string, error) {
	vs, err := s.productRepo.GetLimVars(name, []int{vid})
	if err != nil {
		return models.ProductRedis{}, "", err
	} else if len(vs) != 1 {
		return models.ProductRedis{}, "", fmt.Errorf("different than 1 returned for variant ID: %d", vid)
	}

	return s.productRepo.GetFullProduct(name, vs[0].Handle)
}

func (s *productService) GetProductsByVariantIDs(name string, vids []int) (map[int]*models.ProductRedis, error) {
	vs, err := s.productRepo.GetLimVars(name, vids)
	if err != nil {
		return nil, err
	} else if len(vs) != len(vids) {
		return nil, fmt.Errorf("incorrect limited variant count for supplied variant IDs ")
	}

	handles := []string{}
	for _, v := range vs {
		if !slices.Contains(handles, v.Handle) {
			handles = append(handles, v.Handle)
		}
	}

	sl, err := s.productRepo.GetFullProducts(name, handles)
	if err != nil {
		return nil, err
	}

	ret := map[int]*models.ProductRedis{}

	for _, l := range sl {
		ret[l.PK] = l
	}

	return ret, nil
}

func (s *productService) GetProductsMapFromCartLine(name string, cartLines []models.CartLine) (map[int]*models.ProductRedis, error) {
	vids := []int{}

	for _, cl := range cartLines {
		if !cl.IsGiftCard {
			vids = append(vids, cl.VariantID)
		}
	}

	return s.GetProductsByVariantIDs(name, vids)
}

// Logistics error, DB error
func (s *productService) UpdateRatings(dpi *DataPassIn, pid, newRate, oldRate, plusMinus int, tools *config.Tools) {
	if plusMinus != -1 && plusMinus != 0 && plusMinus != 1 {
		emails.AlertGeneralRatingsError(pid, "", dpi.Store, true, errors.New("plusMinus must be -1 (delete), 0 (update), 1 (new)"), "plusMinus must be -1 (delete), 0 (update), 1 (new)", tools)
		return
	}

	if newRate < 1 || newRate > 5 || ((oldRate < 1 || oldRate > 5) && plusMinus == 0) {
		emails.AlertGeneralRatingsError(pid, "", dpi.Store, true, errors.New("ratings must be 1-5 inclusive"), "ratings must be 1-5 inclusive", tools)
		return
	}

	prod, err := s.productRepo.Read(pid)
	if err != nil {
		emails.AlertGeneralRatingsError(pid, "", dpi.Store, false, err, "unable to read product from SQL", tools)
		return
	}

	prodRedis, rdr, err := s.productRepo.GetFullProduct(dpi.Store, prod.Handle)
	if err != nil {
		emails.AlertGeneralRatingsError(pid, "", dpi.Store, false, err, "unable to read product from redis", tools)
		return
	} else if rdr != "" {
		emails.AlertGeneralRatingsError(pid, "", dpi.Store, false, errors.New("product has redirect, no longer active"), "product has redirect, no longer active", tools)
		return
	}

	prodInfo, err := s.productRepo.GetAllProductInfo(dpi.Store)
	if err != nil {
		emails.AlertGeneralRatingsError(pid, "", dpi.Store, false, err, "unable to read product info section from redis", tools)
		return
	}

	if prod.Rating != prodRedis.Rating || prod.RatingCt != prodRedis.RatingCt {
		emails.AlertRatingsMismatch(pid, prod.Handle, prod.Rating, prodRedis.Rating, prod.RatingCt, prodRedis.RatingCt, dpi.Store, tools)
	}

	rate := prodRedis.Rating
	ct := prodRedis.RatingCt

	if plusMinus == 0 {
		rate += float64(newRate - oldRate)
	} else {
		rate += float64(plusMinus * newRate)
	}
	ct += plusMinus

	if ct < 0 {
		ct = 0
	}

	if ct == 0 {
		rate = 0
	} else {
		if rate < 1 {
			rate = 1
		} else if rate > 5 {
			rate = 5
		}
	}

	prod.Rating = rate
	prod.RatingCt = ct
	prodRedis.Rating = rate
	prodRedis.RatingCt = ct

	found := false
	for i, pi := range prodInfo {
		if pi.ID == pid {
			found = true
			pi.AvgRate = rate
			pi.RateCt = ct
			prodInfo[i] = pi
			break
		}
	}

	if !found {
		emails.AlertProductNotInInfo(pid, prod.Handle, dpi.Store, tools)
	}

	if err := s.productRepo.SaveProductInfoInTransaction(dpi.Store, &prodRedis, prodInfo); err != nil {
		emails.AlertGeneralRatingsError(pid, "", dpi.Store, false, err, "unable to save product and product info in transaction redis", tools)
		return
	}

	if err := s.productRepo.Update(*prod); err != nil {
		emails.AlertGeneralRatingsError(pid, "", dpi.Store, false, err, "unable to save product for sql", tools)
		return
	}
}

func (s *productService) ConfirmDraftOrderProducts(dpi *DataPassIn, vids []int) (map[int]bool, bool, error) {
	lvl, err := s.productRepo.GetLimVars(dpi.Store, vids)
	if err != nil {
		return nil, false, err
	}

	result := map[int]bool{}
	anyFalse := false
	variantIDSet := map[int]struct{}{}
	for _, variant := range lvl {
		variantIDSet[variant.VariantID] = struct{}{}
	}

	for _, vid := range vids {
		_, exists := variantIDSet[vid]
		result[vid] = exists
		if !exists {
			anyFalse = true
		}
	}

	return result, anyFalse, nil
}

func (s *productService) RenderComparables(name string, productID int) ([]models.ComparablesRender, error) {
	comps, err := s.productRepo.ReadComparables(productID)
	if err != nil {
		return nil, err
	}

	if len(comps) == 0 {
		return nil, nil
	}

	otherIDs := map[int]struct{}{}
	for _, comp := range comps {
		if comp.PKFKProductID1 == productID {
			otherIDs[comp.PKFKProductID2] = struct{}{}
		} else {
			otherIDs[comp.PKFKProductID1] = struct{}{}
		}
	}

	prodInfo, err := s.productRepo.GetAllProductInfo(name)
	if err != nil {
		return nil, err
	}

	ret := []models.ComparablesRender{}
	for _, pi := range prodInfo {
		if _, ok := otherIDs[pi.ID]; ok {
			ret = append(ret, models.ComparablesRender{
				Handle:    pi.Handle,
				Title:     pi.Title,
				ImageURL:  pi.ImageURL,
				Price:     pi.Price,
				Inventory: pi.Inventory,
				AvgRate:   pi.AvgRate,
				RateCt:    pi.RateCt,
			})
			delete(otherIDs, pi.ID)
		}
	}

	if len(otherIDs) > 0 {
		log.Printf("Unable to locate ids listed as comps within product info for store: %s, product id: %d, ids: ", name, productID)
		for id := range otherIDs {
			log.Printf(" %d,", id)
		}
		log.Printf("\n")
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].AvgRate > ret[j].AvgRate
	})

	return ret, nil
}

func (s *productService) SetInventoryFromOrder(dpi *DataPassIn, decrement map[int]int, handles []string) error {

	prods, err := s.productRepo.GetFullProducts(dpi.Store, handles)
	if err != nil {
		return err
	}

	maxEach := map[string]int{}
	salesInc := map[string]int{}

	for i, p := range prods {
		maxCurrent := 0
		for j, v := range p.Variants {
			if dec, ok := decrement[v.PK]; ok {
				v.Quantity -= dec
				if config.INV_ALWAYS_UP && v.Quantity < 50 {
					rangeRand := config.HIGHER_INV - config.LOWER_INV + 1
					if rangeRand < 0 {
						rangeRand = 0
					}
					v.Quantity = rand.Intn(rangeRand) + config.LOWER_INV
				}
				if v.Quantity < 0 {
					log.Printf("Negative inventory for handle: %s; variant id: %d; store: %s; inventory: %d\n", p.Handle, v.PK, dpi.Store, v.Quantity)
				}
				if salesCurrent, ok := salesInc[p.Handle]; ok {
					salesInc[p.Handle] = salesCurrent + dec
				} else {
					salesInc[p.Handle] = dec
				}
			}
			if v.Quantity > maxCurrent {
				maxCurrent = v.Quantity
			}
			p.Variants[j] = v
		}
		maxEach[p.Handle] = maxCurrent
		prods[i] = p
	}

	productInfo, err := s.productRepo.GetAllProductInfo(dpi.Store)
	if err != nil {
		return err
	}

	modded := false
	for i, pi := range productInfo {
		if maxNew, ok := maxEach[pi.Handle]; ok {
			if maxNew != pi.Inventory {
				pi.Inventory = maxNew
				modded = true
			}
		}
		if salesUp, ok := salesInc[pi.Handle]; ok {
			pi.Sales += salesUp
		}
		productInfo[i] = pi
	}

	errSaveRedis := error(nil)
	if modded {
		errSaveRedis = s.productRepo.SaveProductInfoInTransactionMulti(dpi.Store, prods, productInfo)
	} else {
		errSaveRedis = s.productRepo.SaveProducts(dpi.Store, prods)
	}

	if errSaveRedis != nil {
		return errSaveRedis
	}

	return s.productRepo.DecrementQuantitiesSQL(decrement)
}
