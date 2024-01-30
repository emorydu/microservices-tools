package grompgsql

import (
	"context"
	"gorm.io/gorm"
)

// genericRepository grom generic repository.
type genericRepository[T any] struct {
	db *gorm.DB
}

// NewGenericRepository create new gorm generic repository.
func NewGenericRepository[T any](db *gorm.DB) *genericRepository[T] {
	return &genericRepository[T]{
		db: db,
	}
}

func (r *genericRepository[T]) Add(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(&entity).Error
}

func (r *genericRepository[T]) AddAll(ctx context.Context, entity *[]T) error {
	return r.db.WithContext(ctx).Create(&entity).Error
}

func (r *genericRepository[T]) GetByID(ctx context.Context, id int) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).Model(&entity).Where("id = ? AND is_active = ?", id, true).FirstOrInit(&entity).Error
	if err != nil {
		return nil, err
	}

	return &entity, nil
}

func (r *genericRepository[T]) Get(ctx context.Context, params *T) *T {
	var entity T
	r.db.WithContext(ctx).Where(&params).FirstOrInit(&entity)

	return &entity
}

func (r *genericRepository[T]) GetAll(ctx context.Context) (*[]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return &entities, nil
}

func (r *genericRepository[T]) Where(ctx context.Context, params *T) (*[]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Where(&params).Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return &entities, nil
}

func (r *genericRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(&entity).Error
}

func (r *genericRepository[T]) UpdateAll(ctx context.Context, entities *[]T) error {
	return r.db.WithContext(ctx).Save(&entities).Error
}

func (r *genericRepository[T]) Delete(ctx context.Context, id int) error {
	var entity T

	return r.db.WithContext(ctx).FirstOrInit(&entity).UpdateColumn("is_active", false).Error
}

func (r *genericRepository[T]) SkipTake(ctx context.Context, skip int, take int) (*[]T, error) {
	var entities []T
	err := r.db.WithContext(ctx).Offset(skip).Limit(take).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return &entities, nil
}

func (r *genericRepository[T]) Count(ctx context.Context) int64 {
	var entity T
	var count int64
	r.db.WithContext(ctx).Model(&entity).Count(&count)
	return count
}

func (r *genericRepository[T]) CountWhere(ctx context.Context, params *T) int64 {
	var entity T
	var count int64
	r.db.WithContext(ctx).Model(&entity).Where(&params).Count(&count)
	return count
}
