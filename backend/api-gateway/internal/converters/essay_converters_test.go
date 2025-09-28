package converters

import (
	"testing"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
	reviewPb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMarshalProtoEssayResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    *pb.EssayResponse
		expected gin.H
	}{
		{
			name: "success - converts complete essay response",
			input: &pb.EssayResponse{
				Id:        1,
				Content:   "Test essay content",
				Author:    "testauthor",
				CreatedAt: 1234567890,
			},
			expected: gin.H{
				"id":         int32(1),
				"content":    "Test essay content",
				"author":     "testauthor",
				"created_at": int64(1234567890),
			},
		},
		{
			name: "success - converts essay with zero values",
			input: &pb.EssayResponse{
				Id:        0,
				Content:   "",
				Author:    "",
				CreatedAt: 0,
			},
			expected: gin.H{
				"id":         int32(0),
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
			result := MarshalProtoEssayResponse(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMarshalProtoEssayWithReviewsResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    *pb.EssayWithReviewsResponse
		expected gin.H
	}{
		{
			name: "success - converts essay with reviews",
			input: &pb.EssayWithReviewsResponse{
				Id:        1,
				Content:   "Test essay content",
				Author:    "testauthor",
				CreatedAt: 1234567890,
				Reviews: []*reviewPb.ReviewResponse{
					{
						Id:        1,
						EssayId:   1,
						Rank:      5,
						Content:   "Great essay!",
						Author:    "reviewer1",
						CreatedAt: 1234567891,
					},
					{
						Id:        2,
						EssayId:   1,
						Rank:      4,
						Content:   "Good essay",
						Author:    "reviewer2",
						CreatedAt: 1234567892,
					},
				},
			},
			expected: gin.H{
				"id":         int32(1),
				"content":    "Test essay content",
				"author":     "testauthor",
				"created_at": int64(1234567890),
				"reviews": []gin.H{
					{
						"id":         int32(1),
						"essay_id":   int32(1),
						"rank":       int32(5),
						"content":    "Great essay!",
						"author":     "reviewer1",
						"created_at": int64(1234567891),
					},
					{
						"id":         int32(2),
						"essay_id":   int32(1),
						"rank":       int32(4),
						"content":    "Good essay",
						"author":     "reviewer2",
						"created_at": int64(1234567892),
					},
				},
			},
		},
		{
			name: "success - converts essay with empty reviews",
			input: &pb.EssayWithReviewsResponse{
				Id:        1,
				Content:   "Test essay content",
				Author:    "testauthor",
				CreatedAt: 1234567890,
				Reviews:   []*reviewPb.ReviewResponse{},
			},
			expected: gin.H{
				"id":         int32(1),
				"content":    "Test essay content",
				"author":     "testauthor",
				"created_at": int64(1234567890),
				"reviews":    []gin.H{},
			},
		},
		{
			name: "success - converts essay with nil reviews",
			input: &pb.EssayWithReviewsResponse{
				Id:        1,
				Content:   "Test essay content",
				Author:    "testauthor",
				CreatedAt: 1234567890,
				Reviews:   nil,
			},
			expected: gin.H{
				"id":         int32(1),
				"content":    "Test essay content",
				"author":     "testauthor",
				"created_at": int64(1234567890),
				"reviews":    []gin.H{},
			},
		},
		{
			name: "success - handles nil review in slice",
			input: &pb.EssayWithReviewsResponse{
				Id:        1,
				Content:   "Test essay content",
				Author:    "testauthor",
				CreatedAt: 1234567890,
				Reviews: []*reviewPb.ReviewResponse{
					nil,
					{
						Id:        1,
						EssayId:   1,
						Rank:      5,
						Content:   "Great essay!",
						Author:    "reviewer1",
						CreatedAt: 1234567891,
					},
				},
			},
			expected: gin.H{
				"id":         int32(1),
				"content":    "Test essay content",
				"author":     "testauthor",
				"created_at": int64(1234567890),
				"reviews": []gin.H{
					gin.H{},
					{
						"id":         int32(1),
						"essay_id":   int32(1),
						"rank":       int32(5),
						"content":    "Great essay!",
						"author":     "reviewer1",
						"created_at": int64(1234567891),
					},
				},
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
			result := MarshalProtoEssayWithReviewsResponse(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
