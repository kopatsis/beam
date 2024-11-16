package services

import (
	"beam/config"
	"beam/data/models"
	"beam/data/repositories"
	"beam/data/services/product"
	"fmt"
	"net/url"
)

type ProductService interface {
	AddProduct(product models.Product) error
	GetProductByID(id int) (*models.Product, error)
	UpdateProduct(product models.Product) error
	DeleteProduct(id int) error
	GetAllProductInfo(fromURL url.Values, Mutex *config.AllMutexes, name string) (models.CollectionRender, error)
	GetProduct(Mutex *config.AllMutexes, name, handle, id string) (models.ProductRender, error)
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

func (s *productService) GetProduct(Mutex *config.AllMutexes, name, handle, id string) (models.ProductRender, error) {
	ret := models.ProductRender{}

	product, redir, err := s.productRepo.GetFullProduct(name, handle)
	if err != nil {
		return ret, err
	}
	fmt.Print(product, redir)

	return ret, nil
}
