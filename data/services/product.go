package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/product"
	"fmt"
	"net/url"
	"strconv"
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
			Price:     fmt.Sprintf("%.2f", float64(rprod.Variants[0].Price)/100),
			CompareAt: fmt.Sprintf("%.2f", float64(rprod.Variants[0].CompareAtPrice)/100),
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
		Price:       fmt.Sprintf("%.2f", float64(rprod.Variants[0].Price)/100),
		CompareAt:   fmt.Sprintf("%.2f", float64(rprod.Variants[0].CompareAtPrice)/100),
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
