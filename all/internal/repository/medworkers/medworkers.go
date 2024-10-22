package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"service/all/internal/entity"
	"service/all/internal/repository/config"
	"service/all/internal/repository/mapper"
	"service/all/internal/repository/models"
	"service/all/internal/repository/utils"
)

type MedicalWorkerRepo struct {
	db *gorm.DB
}

func NewRepository(cfg *config.DB) (*MedicalWorkerRepo, error) {
	fmt.Println("Connecting with DSN:", utils.GetDSN(cfg))
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: utils.GetDSN(cfg), //Непонятно, нужен ли указатель
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("create database gorm object: %w", err)
	}
	db.AutoMigrate(&models.MedWorkerInfo{})

	return &MedicalWorkerRepo{
		db: db,
	}, nil
}

func (r *MedicalWorkerRepo) GetMedicalWorkerByID(ctx context.Context, ID int) (*entity.MedicalWorker, error) {
	var worker models.MedWorkerInfo

	query := r.db.WithContext(ctx).
		Model(&models.MedWorkerInfo{}).
		Where("id = ?", ID)
	if err := query.Take(&worker).Error; err != nil {
		if errors.Is(query.Error, gorm.ErrRecordNotFound) {

			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	medworker, err := mapper.ToMedWorkerEntity(&worker)
	if err != nil {
		return nil, err
	}
	return medworker, nil
}

func (r *MedicalWorkerRepo) AddMedicalWorker(ctx context.Context, medworker *entity.MedicalWorker) (int, error) {
	worker, err := mapper.ToMedWorkerModel(medworker)
	if err != nil {
		return 0, err
	}

	if err := r.db.WithContext(ctx).
		Model(&models.MedWorkerInfo{}).
		Create(&worker).
		Error; err != nil {
		return 0, err
	}
	return int(medworker.ID), nil
}

func (r *MedicalWorkerRepo) UpdateMedicalWorker(ctx context.Context, medworker *entity.MedicalWorker) error {
	worker, err := mapper.ToMedWorkerModel(medworker)
	if err != nil {
		return err
	}

	query := r.db.WithContext(ctx).
		Model(&models.MedWorkerInfo{}).
		Where("id = ?", medworker.ID).
		Updates(worker)
	if err := query.Error; err != nil {
		return err
	}

	if query.RowsAffected == 0 {
		return entity.ErrNotFound
	}

	return nil
}

func (r *MedicalWorkerRepo) PatchMedicalWorker(ctx context.Context, medworker *entity.MedicalWorker) error {
	worker, err := mapper.ToMedWorkerModel(medworker)
	if err != nil {
		return err
	}

	query := r.db.WithContext(ctx).
		Model(&models.MedWorkerInfo{}).
		Where("id = ?", medworker.ID).
		Updates(worker)
	if err := query.Error; err != nil {
		return err
	}

	if query.RowsAffected == 0 {
		return entity.ErrNotFound
	}

	return nil
}

func (r *MedicalWorkerRepo) ListMedicalWorkers(ctx context.Context, limit, offset int) ([]*entity.MedicalWorker, int, error) {
	var workers []models.MedWorkerInfo
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.MedWorkerInfo{}).
		Count(&total)
	if err := query.Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&workers).Error; err != nil {
		return nil, 0, err
	}

	entities := make([]*entity.MedicalWorker, len(workers))
	for i, worker := range workers {
		medworker, err := mapper.ToMedWorkerEntity(&worker)
		if err != nil {
			return nil, 0, err
		}
		entities[i] = medworker
	}

	return entities, int(total), nil
}

func (r *MedicalWorkerRepo) DeleteMedicalWorker(ctx context.Context, ID int) error {
	query := r.db.WithContext(ctx).
		Model(&models.MedWorkerInfo{}).
		Where("id = ?", ID).
		Delete(&models.MedWorkerInfo{})
	if err := query.Error; err != nil {
		return err
	}

	if query.RowsAffected == 0 {
		return entity.ErrNotFound
	}

	return nil
}

func (r *MedicalWorkerRepo) GetPatientsByMedWorker(ctx context.Context, medWorkerID uint64) ([]*entity.PatientCard, error) {
	var cards []models.PatientCardInfo
	if err := r.db.WithContext(ctx).
		Where("med_worker_id = ?", medWorkerID).
		Find(&cards).Error; err != nil {
		return nil, err
	}
	patientCards := make([]*entity.PatientCard, len(cards))
	for i := range cards {
		patientCard, err := mapper.PatientCardToEntity(&cards[i])
		if err != nil {
			return nil, err
		}
		patientCards[i] = patientCard
	}

	return patientCards, nil
}
