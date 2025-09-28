package converters

import (
	"testing"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMarshalReviewResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    *pb.ReviewResponse
		expected gin.H
	}{
		{
			name: "success - converts complete review response",
			input: &pb.ReviewResponse{
				Id:        1,
				EssayId:   2,
				Rank:      5,
				Content:   "Excellent essay!",
				Author:    "reviewer1",
				CreatedAt: 1234567890,
			},
			expected: gin.H{
				"id":         int32(1),
				"essay_id":   int32(2),
				"rank":       int32(5),
				"content":    "Excellent essay!",
				"author":     "reviewer1",
				"created_at": int64(1234567890),
			},
		},
		{
			name: "success - converts review with zero values",
			input: &pb.ReviewResponse{
				Id:        0,
				EssayId:   0,
				Rank:      0,
				Content:   "",
				Author:    "",
				CreatedAt: 0,
			},
			expected: gin.H{
				"id":         int32(0),
				"essay_id":   int32(0),
				"rank":       int32(0),
				"content":    "",
				"author":     "",
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
			result := MarshalReviewResponse(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
