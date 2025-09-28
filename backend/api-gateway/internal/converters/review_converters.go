package converters

import (
	"github.com/gin-gonic/gin"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)

func MarshalReviewResponse(r *pb.ReviewResponse) gin.H {
	if r == nil {
		return gin.H{}
	}
	return gin.H{
		"id":         r.Id,
		"essay_id":   r.EssayId,
		"rank":       r.Rank,
		"content":    r.Content,
		"author":     r.Author,
		"created_at": r.CreatedAt,
	}
}
