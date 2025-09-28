package converters

import (
	"github.com/gin-gonic/gin"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
)


func MarshalProtoEssayResponse(e *pb.EssayResponse) gin.H {
	if e == nil {
		return gin.H{}
	}
	return gin.H{
		"id":         e.Id,
		"content":    e.Content,
		"author":     e.Author,
		"created_at": e.CreatedAt,
	}
}

func MarshalProtoEssayWithReviewsResponse(e *pb.EssayWithReviewsResponse) gin.H {
	if e == nil {
		return gin.H{}
	}
	reviews := make([]gin.H, 0, len(e.Reviews))
	for _, review := range e.Reviews {
		reviews = append(reviews, MarshalReviewResponse(review))
	}
	return gin.H{
		"id":         e.Id,
		"content":    e.Content,
		"author":     e.Author,
		"created_at": e.CreatedAt,
		"reviews":    reviews,
	}
}
