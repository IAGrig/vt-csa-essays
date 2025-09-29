package service

import (
	"context"
	"testing"

	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/repository/mocks"
	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	jwtMocks "github.com/IAGrig/vt-csa-essays/backend/shared/jwt/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.UserRegisterRequest
		setupMock      func(*mocks.MockUserRepository)
		expectedResult *pb.UserResponse
		expectedError  error
	}{
		{
			name: "success - registers user successfully",
			input: &pb.UserRegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				expectedRequest := models.UserLoginRequest{
					Username: "testuser",
					Password: "password123",
				}
				expectedUser := models.User{
					ID:       1,
					Username: "testuser",
				}
				mockRepo.On("Add", expectedRequest).Return(expectedUser, nil)
			},
			expectedResult: &pb.UserResponse{
				Id:       1,
				Username: "testuser",
			},
			expectedError: nil,
		},
		{
			name: "error - duplicate user",
			input: &pb.UserRegisterRequest{
				Username: "existinguser",
				Password: "password123",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				expectedRequest := models.UserLoginRequest{
					Username: "existinguser",
					Password: "password123",
				}
				mockRepo.On("Add", expectedRequest).Return(models.User{}, repository.DuplicateErr)
			},
			expectedResult: nil,
			expectedError:  repository.DuplicateErr,
		},
		{
			name: "error - repository error",
			input: &pb.UserRegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				expectedRequest := models.UserLoginRequest{
					Username: "testuser",
					Password: "password123",
				}
				mockRepo.On("Add", expectedRequest).Return(models.User{}, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			mockGenerator := new(jwtMocks.MockTokenGenerator)
			mockParser := new(jwtMocks.MockTokenParser)
			logger := logging.NewEmptyLogger()

			tt.setupMock(mockRepo)

			service := New(mockRepo, mockGenerator, mockParser, logger)
			result, err := service.Register(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Id, result.Id)
				assert.Equal(t, tt.expectedResult.Username, result.Username)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Auth(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.UserLoginRequest
		setupMock      func(*mocks.MockUserRepository, *jwtMocks.MockTokenGenerator)
		expectedResult *pb.AuthTokensResponse
		expectedError  error
	}{
		{
			name: "success - authenticates user and returns tokens",
			input: &pb.UserLoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository, mockGenerator *jwtMocks.MockTokenGenerator) {
				expectedRequest := models.UserLoginRequest{
					Username: "testuser",
					Password: "password123",
				}
				expectedUser := models.User{
					ID:       1,
					Username: "testuser",
				}
				mockRepo.On("Auth", expectedRequest).Return(expectedUser, nil)
				mockGenerator.On("GenerateAccessToken", "testuser").Return("access_token_123", nil)
				mockGenerator.On("GenerateRefreshToken", "testuser").Return("refresh_token_456", nil)
			},
			expectedResult: &pb.AuthTokensResponse{
				AccessToken:  "access_token_123",
				RefreshToken: "refresh_token_456",
			},
			expectedError: nil,
		},
		{
			name: "error - authentication failed",
			input: &pb.UserLoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository, mockGenerator *jwtMocks.MockTokenGenerator) {
				expectedRequest := models.UserLoginRequest{
					Username: "testuser",
					Password: "wrongpassword",
				}
				mockRepo.On("Auth", expectedRequest).Return(models.User{}, repository.AuthErr)
			},
			expectedResult: nil,
			expectedError:  repository.AuthErr,
		},
		{
			name: "error - token generation fails",
			input: &pb.UserLoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository, mockGenerator *jwtMocks.MockTokenGenerator) {
				expectedRequest := models.UserLoginRequest{
					Username: "testuser",
					Password: "password123",
				}
				expectedUser := models.User{
					ID:       1,
					Username: "testuser",
				}
				mockRepo.On("Auth", expectedRequest).Return(expectedUser, nil)
				mockGenerator.On("GenerateAccessToken", "testuser").Return("", assert.AnError)
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			mockGenerator := new(jwtMocks.MockTokenGenerator)
			mockParser := new(jwtMocks.MockTokenParser)
			logger := logging.NewEmptyLogger()

			tt.setupMock(mockRepo, mockGenerator)

			service := New(mockRepo, mockGenerator, mockParser, logger)
			result, err := service.Auth(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.AccessToken, result.AccessToken)
				assert.Equal(t, tt.expectedResult.RefreshToken, result.RefreshToken)
			}

			mockRepo.AssertExpectations(t)
			mockGenerator.AssertExpectations(t)
		})
	}
}

func TestAuthService_GetByUsername(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.GetByUsernameRequest
		setupMock      func(*mocks.MockUserRepository)
		expectedResult *pb.UserResponse
		expectedError  error
	}{
		{
			name: "success - returns user by username",
			input: &pb.GetByUsernameRequest{
				Username: "testuser",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				expectedUser := models.User{
					ID:       1,
					Username: "testuser",
				}
				mockRepo.On("GetByUsername", "testuser").Return(expectedUser, nil)
			},
			expectedResult: &pb.UserResponse{
				Id:       1,
				Username: "testuser",
			},
			expectedError: nil,
		},
		{
			name: "error - user not found",
			input: &pb.GetByUsernameRequest{
				Username: "nonexistent",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetByUsername", "nonexistent").Return(models.User{}, repository.NotFoundErr)
			},
			expectedResult: nil,
			expectedError:  repository.NotFoundErr,
		},
		{
			name: "error - repository error",
			input: &pb.GetByUsernameRequest{
				Username: "testuser",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetByUsername", "testuser").Return(models.User{}, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			mockGenerator := new(jwtMocks.MockTokenGenerator)
			mockParser := new(jwtMocks.MockTokenParser)
			logger := logging.NewEmptyLogger()

			tt.setupMock(mockRepo)

			service := New(mockRepo, mockGenerator, mockParser, logger)
			result, err := service.GetByUsername(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Id, result.Id)
				assert.Equal(t, tt.expectedResult.Username, result.Username)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.RefreshTokenRequest
		setupMock      func(*mocks.MockUserRepository, *jwtMocks.MockTokenParser, *jwtMocks.MockTokenGenerator)
		expectedResult *pb.AuthTokensResponse
		expectedError  error
	}{
		{
			name: "success - refreshes access token",
			input: &pb.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository, mockParser *jwtMocks.MockTokenParser, mockGenerator *jwtMocks.MockTokenGenerator) {
				mockParser.On("GetUsername", "valid_refresh_token", "refresh").Return("testuser", nil)

				expectedUser := models.User{
					ID:       1,
					Username: "testuser",
				}
				mockRepo.On("GetByUsername", "testuser").Return(expectedUser, nil)

				mockGenerator.On("GenerateAccessToken", "testuser").Return("new_access_token", nil)
			},
			expectedResult: &pb.AuthTokensResponse{
				AccessToken:  "new_access_token",
				RefreshToken: "",
			},
			expectedError: nil,
		},
		{
			name: "error - invalid refresh token",
			input: &pb.RefreshTokenRequest{
				RefreshToken: "invalid_token",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository, mockParser *jwtMocks.MockTokenParser, mockGenerator *jwtMocks.MockTokenGenerator) {
				mockParser.On("GetUsername", "invalid_token", "refresh").Return("", assert.AnError)
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
		{
			name: "error - user not found",
			input: &pb.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository, mockParser *jwtMocks.MockTokenParser, mockGenerator *jwtMocks.MockTokenGenerator) {
				mockParser.On("GetUsername", "valid_refresh_token", "refresh").Return("deleteduser", nil)
				mockRepo.On("GetByUsername", "deleteduser").Return(models.User{}, repository.NotFoundErr)
			},
			expectedResult: nil,
			expectedError:  repository.NotFoundErr,
		},
		{
			name: "error - token generation fails",
			input: &pb.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			setupMock: func(mockRepo *mocks.MockUserRepository, mockParser *jwtMocks.MockTokenParser, mockGenerator *jwtMocks.MockTokenGenerator) {
				mockParser.On("GetUsername", "valid_refresh_token", "refresh").Return("testuser", nil)

				expectedUser := models.User{
					ID:       1,
					Username: "testuser",
				}
				mockRepo.On("GetByUsername", "testuser").Return(expectedUser, nil)

				mockGenerator.On("GenerateAccessToken", "testuser").Return("", assert.AnError)
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			mockGenerator := new(jwtMocks.MockTokenGenerator)
			mockParser := new(jwtMocks.MockTokenParser)
			logger := logging.NewEmptyLogger()

			tt.setupMock(mockRepo, mockParser, mockGenerator)

			service := New(mockRepo, mockGenerator, mockParser, logger)
			result, err := service.RefreshToken(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.AccessToken, result.AccessToken)
				assert.Equal(t, tt.expectedResult.RefreshToken, result.RefreshToken)
			}

			mockRepo.AssertExpectations(t)
			mockParser.AssertExpectations(t)
			mockGenerator.AssertExpectations(t)
		})
	}
}

func TestToProtoUserResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    models.User
		expected *pb.UserResponse
	}{
		{
			name: "converts user to proto response",
			input: models.User{
				ID:       1,
				Username: "testuser",
			},
			expected: &pb.UserResponse{
				Id:       1,
				Username: "testuser",
			},
		},
		{
			name: "handles zero values",
			input: models.User{
				ID:       0,
				Username: "",
			},
			expected: &pb.UserResponse{
				Id:       0,
				Username: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toProtoUserResponse(tt.input)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.Username, result.Username)
		})
	}
}
