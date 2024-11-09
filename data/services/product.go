package services

import (
	"beam/data/models"
	"beam/data/repositories"
)

type ProductService interface {
	AddProduct(product models.Product) error
	GetProductByID(id int) (*models.Product, error)
	UpdateProduct(product models.Product) error
	DeleteProduct(id int) error
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
