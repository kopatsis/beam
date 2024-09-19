package services

import (
	"errors"
	"beam/data/models"
	"beam/data/repositories"
)

type ProductService interface {
	AddProduct(product models.Product) error
	GetProductByID(id uint) (*models.Product, error)
	UpdateProduct(product models.Product) error
	DeleteProduct(id uint) error
}

type productService struct {
	productRepo repositories.ProductRepository
}

func NewProductService(productRepo repositories.ProductRepository) ProductService {
	return &productService{productRepo: productRepo}
}

func (s *productService) AddProduct(product models.Product) error {
	if product.Name == "" {
		return errors.New("product name cannot be empty")
	}
	return s.productRepo.Create(product)
}

func (s *productService) GetProductByID(id uint) (*models.Product, error) {
	return s.productRepo.FindByID(id)
}

func (s *productService) UpdateProduct(product models.Product) error {
	existingProduct, err := s.productRepo.FindByID(product.ID)
	if err != nil || existingProduct == nil {
		return errors.New("product not found")
	}
	return s.productRepo.Update(product)
}

func (s *productService) DeleteProduct(id uint) error {
	return s.productRepo.Delete(id)
}
