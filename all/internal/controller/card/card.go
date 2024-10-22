package controller

import (
	"context"
	"errors"
	pb "service/all/api"
	"service/all/internal/entity"
	"service/all/internal/usecase"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedMedCardServer
	cardUseCase usecase.Card
	logger      *zap.Logger
}

func NewServer(cardUseCase usecase.Card, logger *zap.Logger) *Server {
	return &Server{
		cardUseCase: cardUseCase,
		logger:      logger,
	}
}

func (s *Server) GetCards(ctx context.Context, request *pb.GetCardsRequest) (*pb.GetCardsResponse, error) {
	s.logger.Info("[Request] New request", zap.Any("data", request))
	limit := int(request.GetLimit())
	offset := int(request.GetOffset())

	cardList, err := s.cardUseCase.GetCards(ctx, limit, offset)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "Cards were not found")
		}
		s.logger.Error("Failed to fetch cards", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to fetch cards")
	}

	response := &pb.GetCardsResponse{
		Count:   uint64(cardList.Count),
		Results: []*pb.Card{},
	}
	for _, cardInfo := range cardList.Cards {
		cardResponse := &pb.Card{
			Id:              uint64(cardInfo.Card.ID),
			AppointmentTime: cardInfo.Card.AppointmentTime,
			HasNodules:      cardInfo.Card.HasNodules,
			Diagnosis:       cardInfo.Card.Diagnosis,
			Patient: &pb.Patient{
				Id:            cardInfo.Patient.ID,
				FirstName:     cardInfo.Patient.FirstName,
				LastName:      cardInfo.Patient.LastName,
				FatherName:    cardInfo.Patient.FatherName,
				MedicalPolicy: cardInfo.Patient.MedicalPolicy,
				Email:         cardInfo.Patient.Email,
				IsActive:      cardInfo.Patient.IsActive,
			},
			MedWorkerId: cardInfo.Card.MedWorkerID,
		}
		response.Results = append(response.Results, cardResponse)
	}

	return response, nil
}

func (s *Server) PostCard(ctx context.Context, request *pb.PostCardRequest) (*pb.PostCardResponse, error) {
	cardInfo := &entity.PatientInformation{
		Patient: &entity.Patient{
			ID:            request.Patient.Id,
			FirstName:     request.Patient.FirstName,
			LastName:      request.Patient.LastName,
			FatherName:    request.Patient.FatherName,
			MedicalPolicy: request.Patient.MedicalPolicy,
			Email:         request.Patient.Email,
			IsActive:      request.Patient.IsActive,
		},
		Card: &entity.PatientCard{
			HasNodules:  request.HasNodules,
			Diagnosis:   request.Diagnosis,
			PatientID:   request.Patient.Id,
			MedWorkerID: request.MedworkerId,
		},
	}
	err := s.cardUseCase.PostCard(ctx, cardInfo)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return nil, nil
}

func (s *Server) GetCardByID(ctx context.Context, request *pb.GetCardByIDRequest) (*pb.GetCardByIDResponse, error) {
	CardInfo, err := s.cardUseCase.GetCardByID(ctx, uint64(request.Id))
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	response := &pb.GetCardByIDResponse{
		Postcard: &pb.Card{
			Id:              uint64(CardInfo.Card.ID),
			AppointmentTime: CardInfo.Card.AppointmentTime,
			HasNodules:      CardInfo.Card.HasNodules,
			Diagnosis:       CardInfo.Card.Diagnosis,
			Patient: &pb.Patient{
				Id:            CardInfo.Patient.ID,
				FirstName:     CardInfo.Patient.FirstName,
				LastName:      CardInfo.Patient.LastName,
				FatherName:    CardInfo.Patient.FatherName,
				MedicalPolicy: CardInfo.Patient.MedicalPolicy,
				Email:         CardInfo.Patient.Email,
				IsActive:      CardInfo.Patient.IsActive,
			},
			MedWorkerId: CardInfo.Card.MedWorkerID,
		},
	}
	return response, nil

}

func (s *Server) PutCard(ctx context.Context, request *pb.PutCardRequest) (*pb.PutCardResponse, error) {
	Card := &entity.PatientCard{
		ID:          request.Id,
		HasNodules:  request.HasNodules,
		Diagnosis:   request.Diagnosis,
		PatientID:   request.PatientId,
		MedWorkerID: request.MedworkerId,
	}
	err := s.cardUseCase.PutCard(ctx, Card)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return nil, nil
}

func (s *Server) DeleteCard(ctx context.Context, request *pb.DeleteCardRequest) (*pb.DeleteCardResponse, error) {
	s.logger.Info("[Request] Delete card", zap.Any("request", request))

	err := s.cardUseCase.DeleteCard(ctx, request.Id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "Card not found")
		}
		s.logger.Error("Failed to delete card", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to delete card")
	}

	return &pb.DeleteCardResponse{}, nil
}
