package service

import (
	"github.com/IAGrig/vt-csa-essays/backend/review-service/internal/models"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)


func toProtoReviewResponse(r models.Review) *pb.ReviewResponse {
	var createdAt int64
	if !r.CreatedAt.IsZero() {
		createdAt = r.CreatedAt.Unix()
	}

	return &pb.ReviewResponse{
		Id:		int32(r.ID),
		EssayId:   int32(r.EssayId),
		Rank:	  int32(r.Rank),
		Content:   r.Content,
		Author:	r.Author,
		CreatedAt: createdAt,
	}
}
