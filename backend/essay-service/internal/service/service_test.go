package service

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/repository/mocks"
	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
	reviewPb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type MockReviewClient struct {
	mock.Mock
}

func (m *MockReviewClient) Add(ctx context.Context, in *reviewPb.ReviewAddRequest, opts ...grpc.CallOption) (*reviewPb.ReviewResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*reviewPb.ReviewResponse), args.Error(1)
}

func (m *MockReviewClient) GetAllReviews(ctx context.Context, in *reviewPb.EmptyRequest, opts ...grpc.CallOption) (reviewPb.ReviewService_GetAllReviewsClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(reviewPb.ReviewService_GetAllReviewsClient), args.Error(1)
}

func (m *MockReviewClient) GetByEssayId(ctx context.Context, in *reviewPb.GetByEssayIdRequest, opts ...grpc.CallOption) (reviewPb.ReviewService_GetByEssayIdClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(reviewPb.ReviewService_GetByEssayIdClient), args.Error(1)
}

func (m *MockReviewClient) RemoveById(ctx context.Context, in *reviewPb.RemoveByIdRequest, opts ...grpc.CallOption) (*reviewPb.ReviewResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*reviewPb.ReviewResponse), args.Error(1)
}

type MockReviewStream struct {
	mock.Mock
	reviews []*reviewPb.ReviewResponse
	current int
}

func (m *MockReviewStream) Recv() (*reviewPb.ReviewResponse, error) {
	args := m.Called()

	if m.current < len(m.reviews) {
		review := m.reviews[m.current]
		m.current++
		return review, args.Error(0)
	}

	return nil, args.Error(0)
}

func (m *MockReviewStream) Header() (metadata.MD, error) {
	args := m.Called()
	return args.Get(0).(metadata.MD), args.Error(1)
}

func (m *MockReviewStream) Trailer() metadata.MD {
	args := m.Called()
	return args.Get(0).(metadata.MD)
}

func (m *MockReviewStream) CloseSend() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockReviewStream) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *MockReviewStream) SendMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockReviewStream) RecvMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

type MinimalServerStream struct {
	ctx          context.Context
	sentMessages []*pb.EssayResponse
	sendError    error
}

func (m *MinimalServerStream) Send(msg *pb.EssayResponse) error {
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

func TestEssayService_Add(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.EssayAddRequest
		setupMock      func(*mocks.MockEssayRepository)
		expectedResult *pb.EssayResponse
		expectedError  error
	}{
		{
			name: "success - adds essay successfully",
			input: &pb.EssayAddRequest{
				Content: "Test essay content",
				Author:  "testuser",
			},
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				expectedRequest := models.EssayRequest{
					Content: "Test essay content",
					Author:  "testuser",
				}
				expectedEssay := models.Essay{
					ID:      1,
					Content: "Test essay content",
					Author:  "testuser",
				}
				mockRepo.On("Add", expectedRequest).Return(expectedEssay, nil)
			},
			expectedResult: &pb.EssayResponse{
				Id:      1,
				Content: "Test essay content",
				Author:  "testuser",
			},
			expectedError: nil,
		},
		{
			name: "error - duplicate essay",
			input: &pb.EssayAddRequest{
				Content: "Duplicate content",
				Author:  "testuser",
			},
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				expectedRequest := models.EssayRequest{
					Content: "Duplicate content",
					Author:  "testuser",
				}
				mockRepo.On("Add", expectedRequest).Return(models.Essay{}, repository.DuplicateErr)
			},
			expectedResult: nil,
			expectedError:  repository.DuplicateErr,
		},
		{
			name: "error - repository error",
			input: &pb.EssayAddRequest{
				Content: "Test content",
				Author:  "testuser",
			},
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				expectedRequest := models.EssayRequest{
					Content: "Test content",
					Author:  "testuser",
				}
				mockRepo.On("Add", expectedRequest).Return(models.Essay{}, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEssayRepository)
			mockReviewClient := new(MockReviewClient)
			tt.setupMock(mockRepo)

			logger := logging.NewEmptyLogger()
			service := New(mockRepo, mockReviewClient, logger)
			result, err := service.Add(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Id, result.Id)
				assert.Equal(t, tt.expectedResult.Content, result.Content)
				assert.Equal(t, tt.expectedResult.Author, result.Author)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEssayService_GetAllEssays(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockEssayRepository)
		expectedCount  int
		sendError      error
		expectedError  bool
	}{
		{
			name: "success - streams all essays",
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				essays := []models.Essay{
					{ID: 1, Content: "Essay 1", Author: "user1"},
					{ID: 2, Content: "Essay 2", Author: "user2"},
				}
				mockRepo.On("GetAllEssays").Return(essays, nil)
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "success - streams empty list when no essays",
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				mockRepo.On("GetAllEssays").Return([]models.Essay{}, nil)
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "error - repository returns error",
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				mockRepo.On("GetAllEssays").Return([]models.Essay{}, assert.AnError)
			},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name: "error - stream send fails",
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				essays := []models.Essay{
					{ID: 1, Content: "Essay 1", Author: "user1"},
				}
				mockRepo.On("GetAllEssays").Return(essays, nil)
			},
			sendError:     assert.AnError,
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEssayRepository)
			mockReviewClient := new(MockReviewClient)
			tt.setupMock(mockRepo)

			stream := &MinimalServerStream{
				ctx:       context.Background(),
				sendError: tt.sendError,
			}

			logger := logging.NewEmptyLogger()
			service := New(mockRepo, mockReviewClient, logger)
			err := service.GetAllEssays(&pb.EmptyRequest{}, stream)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, stream.sentMessages, tt.expectedCount)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEssayService_GetByAuthorName(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.GetByAuthorNameRequest
		setupMock      func(*mocks.MockEssayRepository, *MockReviewClient, *MockReviewStream)
		expectedResult *pb.EssayWithReviewsResponse
		expectedError  error
	}{
		{
			name: "success - returns essay with reviews",
			input: &pb.GetByAuthorNameRequest{Authorname: "testuser"},
			setupMock: func(mockRepo *mocks.MockEssayRepository, mockReviewClient *MockReviewClient, mockStream *MockReviewStream) {
				expectedEssay := models.Essay{
					ID:      1,
					Content: "Test essay",
					Author:  "testuser",
				}
				mockRepo.On("GetByAuthorName", "testuser").Return(expectedEssay, nil)

				reviewRequest := &reviewPb.GetByEssayIdRequest{EssayId: 1}
				mockReviewClient.On("GetByEssayId", mock.Anything, reviewRequest, mock.Anything).Return(mockStream, nil)

				reviews := []*reviewPb.ReviewResponse{
					{Id: 1, EssayId: 1, Rank: 5, Content: "Great essay", Author: "reviewer1"},
					{Id: 2, EssayId: 1, Rank: 4, Content: "Good essay", Author: "reviewer2"},
				}
				mockStream.reviews = reviews

				mockStream.On("Recv").Return(nil).Once()
				mockStream.On("Recv").Return(nil).Once()
				mockStream.On("Recv").Return(io.EOF).Once()
			},
			expectedResult: &pb.EssayWithReviewsResponse{
				Id:      1,
				Content: "Test essay",
				Author:  "testuser",
				Reviews: []*reviewPb.ReviewResponse{
					{Id: 1, EssayId: 1, Rank: 5, Content: "Great essay", Author: "reviewer1"},
					{Id: 2, EssayId: 1, Rank: 4, Content: "Good essay", Author: "reviewer2"},
				},
			},
			expectedError: nil,
		},
		{
			name: "error - essay not found",
			input: &pb.GetByAuthorNameRequest{Authorname: "nonexistent"},
			setupMock: func(mockRepo *mocks.MockEssayRepository, mockReviewClient *MockReviewClient, mockStream *MockReviewStream) {
				mockRepo.On("GetByAuthorName", "nonexistent").Return(models.Essay{}, repository.EssayNotFoundErr)
			},
			expectedResult: nil,
			expectedError:  repository.EssayNotFoundErr,
		},
		{
			name: "error - review client fails",
			input: &pb.GetByAuthorNameRequest{Authorname: "testuser"},
			setupMock: func(mockRepo *mocks.MockEssayRepository, mockReviewClient *MockReviewClient, mockStream *MockReviewStream) {
				expectedEssay := models.Essay{
					ID:      1,
					Content: "Test essay",
					Author:  "testuser",
				}
				mockRepo.On("GetByAuthorName", "testuser").Return(expectedEssay, nil)

				reviewRequest := &reviewPb.GetByEssayIdRequest{EssayId: 1}
				mockReviewClient.On("GetByEssayId", mock.Anything, reviewRequest, mock.Anything).Return((*MockReviewStream)(nil), errors.New("review service unavailable"))
			},
			expectedResult: nil,
			expectedError:  errors.New("failed to get reviews stream from review service: review service unavailable"),
		},
		{
			name: "error - review stream fails",
			input: &pb.GetByAuthorNameRequest{Authorname: "testuser"},
			setupMock: func(mockRepo *mocks.MockEssayRepository, mockReviewClient *MockReviewClient, mockStream *MockReviewStream) {
				expectedEssay := models.Essay{
					ID:      1,
					Content: "Test essay",
					Author:  "testuser",
				}
				mockRepo.On("GetByAuthorName", "testuser").Return(expectedEssay, nil)

				reviewRequest := &reviewPb.GetByEssayIdRequest{EssayId: 1}
				mockReviewClient.On("GetByEssayId", mock.Anything, reviewRequest, mock.Anything).Return(mockStream, nil)

				mockStream.On("Recv").Return(assert.AnError).Once()
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
		{
			name: "success - essay with no reviews",
			input: &pb.GetByAuthorNameRequest{Authorname: "testuser"},
			setupMock: func(mockRepo *mocks.MockEssayRepository, mockReviewClient *MockReviewClient, mockStream *MockReviewStream) {
				expectedEssay := models.Essay{
					ID:      1,
					Content: "Test essay",
					Author:  "testuser",
				}
				mockRepo.On("GetByAuthorName", "testuser").Return(expectedEssay, nil)

				reviewRequest := &reviewPb.GetByEssayIdRequest{EssayId: 1}
				mockReviewClient.On("GetByEssayId", mock.Anything, reviewRequest, mock.Anything).Return(mockStream, nil)

				mockStream.On("Recv").Return(io.EOF).Once()
			},
			expectedResult: &pb.EssayWithReviewsResponse{
				Id:      1,
				Content: "Test essay",
				Author:  "testuser",
				Reviews: []*reviewPb.ReviewResponse{},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEssayRepository)
			mockReviewClient := new(MockReviewClient)
			mockReviewStream := new(MockReviewStream)
			tt.setupMock(mockRepo, mockReviewClient, mockReviewStream)

			logger := logging.NewEmptyLogger()
			service := New(mockRepo, mockReviewClient, logger)
			result, err := service.GetByAuthorName(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Id, result.Id)
				assert.Equal(t, tt.expectedResult.Content, result.Content)
				assert.Equal(t, tt.expectedResult.Author, result.Author)
				assert.Len(t, result.Reviews, len(tt.expectedResult.Reviews))

				for i, expectedReview := range tt.expectedResult.Reviews {
					assert.Equal(t, expectedReview.Id, result.Reviews[i].Id)
					assert.Equal(t, expectedReview.Content, result.Reviews[i].Content)
					assert.Equal(t, expectedReview.Author, result.Reviews[i].Author)
				}
			}

			mockRepo.AssertExpectations(t)
			mockReviewClient.AssertExpectations(t)
			mockReviewStream.AssertExpectations(t)
		})
	}
}

func TestEssayService_RemoveByAuthorName(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.RemoveByAuthorNameRequest
		setupMock      func(*mocks.MockEssayRepository)
		expectedResult *pb.EssayResponse
		expectedError  error
	}{
		{
			name: "success - removes essay by author name",
			input: &pb.RemoveByAuthorNameRequest{Authorname: "testuser"},
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				expectedEssay := models.Essay{
					ID:      1,
					Content: "Deleted essay",
					Author:  "testuser",
				}
				mockRepo.On("RemoveByAuthorName", "testuser").Return(expectedEssay, nil)
			},
			expectedResult: &pb.EssayResponse{
				Id:      1,
				Content: "Deleted essay",
				Author:  "testuser",
			},
			expectedError: nil,
		},
		{
			name: "error - essay not found",
			input: &pb.RemoveByAuthorNameRequest{Authorname: "nonexistent"},
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				mockRepo.On("RemoveByAuthorName", "nonexistent").Return(models.Essay{}, repository.EssayNotFoundErr)
			},
			expectedResult: nil,
			expectedError:  repository.EssayNotFoundErr,
		},
		{
			name: "error - repository error",
			input: &pb.RemoveByAuthorNameRequest{Authorname: "testuser"},
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				mockRepo.On("RemoveByAuthorName", "testuser").Return(models.Essay{}, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEssayRepository)
			mockReviewClient := new(MockReviewClient)
			tt.setupMock(mockRepo)

			logger := logging.NewEmptyLogger()
			service := New(mockRepo, mockReviewClient, logger)
			result, err := service.RemoveByAuthorName(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Id, result.Id)
				assert.Equal(t, tt.expectedResult.Content, result.Content)
				assert.Equal(t, tt.expectedResult.Author, result.Author)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEssayService_SearchByContent(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.SearchByContentRequest
		setupMock      func(*mocks.MockEssayRepository)
		expectedCount  int
		sendError      error
		expectedError  bool
	}{
		{
			name: "success - streams search results",
			input: &pb.SearchByContentRequest{Content: "search term"},
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				essays := []models.Essay{
					{ID: 1, Content: "Essay with search term", Author: "user1"},
					{ID: 2, Content: "Another essay with search term", Author: "user2"},
				}
				mockRepo.On("SearchByContent", "search term").Return(essays, nil)
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "success - no search results",
			input: &pb.SearchByContentRequest{Content: "nonexistent"},
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				mockRepo.On("SearchByContent", "nonexistent").Return([]models.Essay{}, nil)
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "error - repository returns error",
			input: &pb.SearchByContentRequest{Content: "search term"},
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				mockRepo.On("SearchByContent", "search term").Return([]models.Essay{}, assert.AnError)
			},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name: "error - stream send fails",
			input: &pb.SearchByContentRequest{Content: "search term"},
			setupMock: func(mockRepo *mocks.MockEssayRepository) {
				essays := []models.Essay{
					{ID: 1, Content: "Essay with search term", Author: "user1"},
				}
				mockRepo.On("SearchByContent", "search term").Return(essays, nil)
			},
			sendError:     assert.AnError,
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEssayRepository)
			mockReviewClient := new(MockReviewClient)
			tt.setupMock(mockRepo)

			stream := &MinimalServerStream{
				ctx:       context.Background(),
				sendError: tt.sendError,
			}

			logger := logging.NewEmptyLogger()
			service := New(mockRepo, mockReviewClient, logger)
			err := service.SearchByContent(tt.input, stream)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, stream.sentMessages, tt.expectedCount)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestConverters(t *testing.T) {
	tests := []struct {
		name     string
		input    models.Essay
		expected *pb.EssayResponse
	}{
		{
			name: "converts essay to proto response",
			input: models.Essay{
				ID:      1,
				Content: "Test content",
				Author:  "test author",
			},
			expected: &pb.EssayResponse{
				Id:      1,
				Content: "Test content",
				Author:  "test author",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toProtoEssayResponse(tt.input)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.Content, result.Content)
			assert.Equal(t, tt.expected.Author, result.Author)
		})
	}
}
