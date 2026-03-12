package service

import (
	"context"
	"errors"

	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/pkg/pinyin"
)

var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryCodeExists    = errors.New("category code already exists")
	ErrCategoryHasPackages   = errors.New("category has packages, cannot delete")
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   *int   `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

func (s *CategoryService) Create(ctx context.Context, req *CreateCategoryRequest) (*model.Category, error) {
	// 检查名称是否已存在
	existing, _ := s.categoryRepo.GetByName(ctx, req.Name)
	if existing != nil {
		return nil, ErrCategoryAlreadyExists
	}

	// 生成拼音代码
	code := pinyin.GenerateCode(req.Name)

	// 检查代码是否已存在（理论上不会，因为名称唯一）
	if exists, _ := s.categoryRepo.ExistsByCode(ctx, code); exists {
		// 如果冲突，添加时间戳后缀
		code = code + "_" + "V2"
	}

	category := &model.Category{
		Name:        req.Name,
		Code:        code,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		IsActive:    true,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) Update(ctx context.Context, id uint, req *UpdateCategoryRequest) (*model.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCategoryNotFound
	}

	// 如果要更新名称，需要重新生成代码
	if req.Name != "" && req.Name != category.Name {
		// 检查新名称是否已存在
		existing, _ := s.categoryRepo.GetByName(ctx, req.Name)
		if existing != nil && existing.ID != id {
			return nil, ErrCategoryAlreadyExists
		}
		category.Name = req.Name
		category.Code = pinyin.GenerateCode(req.Name)
	}

	if req.Description != "" {
		category.Description = req.Description
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) Delete(ctx context.Context, id uint) error {
	count, err := s.categoryRepo.CountPackagesByCategory(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrCategoryHasPackages
	}
	return s.categoryRepo.Delete(ctx, id)
}

func (s *CategoryService) GetByID(ctx context.Context, id uint) (*model.Category, error) {
	return s.categoryRepo.GetByID(ctx, id)
}

func (s *CategoryService) GetByCode(ctx context.Context, code string) (*model.Category, error) {
	return s.categoryRepo.GetByCode(ctx, code)
}

func (s *CategoryService) List(ctx context.Context, page, pageSize int) ([]model.Category, int64, error) {
	return s.categoryRepo.List(ctx, page, pageSize)
}

func (s *CategoryService) ListActive(ctx context.Context) ([]model.Category, error) {
	return s.categoryRepo.ListActive(ctx)
}
