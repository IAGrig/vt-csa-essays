package service

import (
	"context"
	"testing"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

type MinimalServerStream struct {
	ctx          context.Context
	sentMessages []*pb.ReviewResponse
	sendError    error
}

func (m *MinimalServerStream) Send(msg *pb.ReviewResponse) error {
	if m.sendError != nil {
		return m.sendError
	}
	m.sentMessages = append(m.sentMessages, msg)
	return nil
}

func (m *MinimalServerStream) Context() context.Context {
	return m.ctx
}

func (m *MinimalServerStream) SetHeader(md metadata.MD) error { return nil }
func (m *MinimalServerStream) SendHeader(md metadata.MD) error { return nil }
func (m *MinimalServerStream) SetTrailer(md metadata.MD)      {}
func (m *MinimalServerStream) SendMsg(interface{}) error      { return nil }
func (m *MinimalServerStream) RecvMsg(interface{}) error      { return nil }

func TestReviewService_Add(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.ReviewAddRequest
		setupMock      func(*mocks.MockReviewRepository)
		expectedResult *pb.ReviewResponse
		expectedError  bool
	}{
		{
			name: "success - adds review successfully",
			input: &pb.ReviewAddRequest{
				EssayId: 1,
				Rank:    1,
				Content: "Excellent essay",
				Author:  "reviewer1",
			},
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				expectedRequest := models.ReviewRequest{
					EssayId: 1,
					Rank:    1,
					Content: "Excellent essay",
					Author:  "reviewer1",
				}
				expectedReview := models.Review{
					ID:      1,
					EssayId: 1,
					Rank:    1,
					Content: "Excellent essay",
					Author:  "reviewer1",
				}
				mockRepo.On("Add", expectedRequest).Return(expectedReview, nil)
			},
			expectedResult: &pb.ReviewResponse{
				Id:      1,
				EssayId: 1,
				Rank:    1,
				Content: "Excellent essay",
				Author:  "reviewer1",
			},
			expectedError: false,
		},
		{
			name: "error - repository returns error",
			input: &pb.ReviewAddRequest{
				EssayId: 1,
				Rank:    1,
				Content: "Excellent essay",
				Author:  "reviewer1",
			},
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				expectedRequest := models.ReviewRequest{
					EssayId: 1,
					Rank:    1,
					Content: "Excellent essay",
					Author:  "reviewer1",
				}
				mockRepo.On("Add", expectedRequest).Return(models.Review{}, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockReviewRepository)
			tt.setupMock(mockRepo)

			service := New(mockRepo)
			result, err := service.Add(context.Background(), tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Id, result.Id)
				assert.Equal(t, tt.expectedResult.EssayId, result.EssayId)
				assert.Equal(t, tt.expectedResult.Rank, result.Rank)
				assert.Equal(t, tt.expectedResult.Content, result.Content)
				assert.Equal(t, tt.expectedResult.Author, result.Author)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestReviewService_GetAllReviews(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockReviewRepository)
		expectedCount  int
		sendError      error
		expectedError  bool
	}{
		{
			name: "success - streams all reviews",
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				reviews := []models.Review{
					{ID: 1, EssayId: 1, Rank: 1, Content: "Great essay", Author: "reviewer1"},
					{ID: 2, EssayId: 2, Rank: 2, Content: "Good essay", Author: "reviewer2"},
				}
				mockRepo.On("GetAllReviews").Return(reviews, nil)
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "success - streams empty list when no reviews",
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				mockRepo.On("GetAllReviews").Return([]models.Review{}, nil)
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "error - repository returns error",
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				mockRepo.On("GetAllReviews").Return([]models.Review{}, assert.AnError)
			},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name: "error - stream send fails",
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				reviews := []models.Review{
					{ID: 1, EssayId: 1, Rank: 1, Content: "Great essay", Author: "reviewer1"},
				}
				mockRepo.On("GetAllReviews").Return(reviews, nil)
			},
			sendError:     assert.AnError,
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockReviewRepository)
			tt.setupMock(mockRepo)

			stream := &MinimalServerStream{
				ctx:       context.Background(),
				sendError: tt.sendError,
			}

			service := New(mockRepo)
			err := service.GetAllReviews(&pb.EmptyRequest{}, stream)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, stream.sentMessages, tt.expectedCount)

			// refactor later
			if tt.expectedCount > 0 {
				assert.Equal(t, int32(1), stream.sentMessages[0].Id)
				assert.Equal(t, "Great essay", stream.sentMessages[0].Content)
				if tt.expectedCount > 1 {
					assert.Equal(t, int32(2), stream.sentMessages[1].Id)
					assert.Equal(t, "Good essay", stream.sentMessages[1].Content)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestReviewService_GetByEssayId(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.GetByEssayIdRequest
		setupMock      func(*mocks.MockReviewRepository)
		expectedCount  int
		sendError      error
		expectedError  bool
	}{
		{
			name:  "success - streams reviews for essay",
			input: &pb.GetByEssayIdRequest{EssayId: 1},
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				reviews := []models.Review{
					{ID: 1, EssayId: 1, Rank: 1, Content: "Review 1", Author: "reviewer1"},
					{ID: 2, EssayId: 1, Rank: 2, Content: "Review 2", Author: "reviewer2"},
				}
				mockRepo.On("GetByEssayId", 1).Return(reviews, nil)
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:  "success - no reviews for essay",
			input: &pb.GetByEssayIdRequest{EssayId: 999},
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				mockRepo.On("GetByEssayId", 999).Return([]models.Review{}, nil)
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name:  "error - repository returns error",
			input: &pb.GetByEssayIdRequest{EssayId: 1},
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				mockRepo.On("GetByEssayId", 1).Return([]models.Review{}, assert.AnError)
			},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name:  "error - stream send fails",
			input: &pb.GetByEssayIdRequest{EssayId: 1},
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				reviews := []models.Review{
					{ID: 1, EssayId: 1, Rank: 1, Content: "Review 1", Author: "reviewer1"},
				}
				mockRepo.On("GetByEssayId", 1).Return(reviews, nil)
			},
			sendError:     assert.AnError,
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockReviewRepository)
			tt.setupMock(mockRepo)

			stream := &MinimalServerStream{
				ctx:       context.Background(),
				sendError: tt.sendError,
			}

			service := New(mockRepo)
			err := service.GetByEssayId(tt.input, stream)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, stream.sentMessages, tt.expectedCount)

			for _, msg := range stream.sentMessages {
				assert.Equal(t, tt.input.EssayId, msg.EssayId)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestReviewService_RemoveById(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.RemoveByIdRequest
		setupMock      func(*mocks.MockReviewRepository)
		expectedResult *pb.ReviewResponse
		expectedError  bool
	}{
		{
			name:  "success - removes review by id",
			input: &pb.RemoveByIdRequest{Id: 1},
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				review := models.Review{
					ID:      1,
					EssayId: 1,
					Rank:    1,
					Content: "Deleted review",
					Author:  "reviewer1",
				}
				mockRepo.On("RemoveById", 1).Return(review, nil)
			},
			expectedResult: &pb.ReviewResponse{
				Id:      1,
				EssayId: 1,
				Rank:    1,
				Content: "Deleted review",
				Author:  "reviewer1",
			},
			expectedError: false,
		},
		{
			name:  "error - review not found",
			input: &pb.RemoveByIdRequest{Id: 999},
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				mockRepo.On("RemoveById", 999).Return(models.Review{}, repository.ReviewNotFoundErr)
			},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:  "error - repository returns error",
			input: &pb.RemoveByIdRequest{Id: 1},
			setupMock: func(mockRepo *mocks.MockReviewRepository) {
				mockRepo.On("RemoveById", 1).Return(models.Review{}, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockReviewRepository)
			tt.setupMock(mockRepo)

			service := New(mockRepo)
			result, err := service.RemoveById(context.Background(), tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Id, result.Id)
				assert.Equal(t, tt.expectedResult.EssayId, result.EssayId)
				assert.Equal(t, tt.expectedResult.Rank, result.Rank)
				assert.Equal(t, tt.expectedResult.Content, result.Content)
				assert.Equal(t, tt.expectedResult.Author, result.Author)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestToProtoReviewResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    models.Review
		expected *pb.ReviewResponse
	}{
		{
			name: "converts review to proto response",
			input: models.Review{
				ID:      1,
				EssayId: 2,
				Rank:    1,
				Content: "Test content",
				Author:  "test author",
			},
			expected: &pb.ReviewResponse{
				Id:      1,
				EssayId: 2,
				Rank:    1,
				Content: "Test content",
				Author:  "test author",
			},
		},
		{
			name: "handles zero values",
			input: models.Review{
				ID:      0,
				EssayId: 0,
				Rank:    0,
				Content: "",
				Author:  "",
			},
			expected: &pb.ReviewResponse{
				Id:      0,
				EssayId: 0,
				Rank:    0,
				Content: "",
				Author:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toProtoReviewResponse(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
