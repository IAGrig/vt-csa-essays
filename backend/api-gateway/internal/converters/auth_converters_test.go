package converters

import (
	"testing"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMarshalProtoUserResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    *pb.UserResponse
		expected gin.H
	}{
		{
			name: "success - converts complete user response",
			input: &pb.UserResponse{
				Id:        1,
				Username:  "testuser",
				CreatedAt: 1234567890,
			},
			expected: gin.H{
				"id":         int32(1),
				"username":   "testuser",
				"created_at": int64(1234567890),
			},
		},
		{
			name: "success - converts user with zero values",
			input: &pb.UserResponse{
				Id:        0,
				Username:  "",
				CreatedAt: 0,
			},
			expected: gin.H{
				"id":         int32(0),
				"username":   "",
				"created_at": int64(0),
			},
		},
		{
			name:     "success - handles nil input",
			input:    nil,
			expected: gin.H{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MarshalProtoUserResponse(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
