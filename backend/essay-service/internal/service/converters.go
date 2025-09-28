package service

import (
	"github.com/IAGrig/vt-csa-essays/backend/essay-service/internal/models"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/essay"
	reviewPb "github.com/IAGrig/vt-csa-essays/backend/proto/review"
)


func toProtoEssayResponse(e models.Essay) *pb.EssayResponse {
	return &pb.EssayResponse{
		Id:        int32(e.ID),
		Content:   e.Content,
		Author:    e.Author,
		CreatedAt: e.CreatedAt.Unix(),
	}
}

func toProtoEssayWithReviewsResponse(e models.Essay, reviews []*reviewPb.ReviewResponse) *pb.EssayWithReviewsResponse {
	return &pb.EssayWithReviewsResponse{
		Id: int32(e.ID),
		Content: e.Content,
		Author: e.Author,
		CreatedAt: e.CreatedAt.Unix(),
		Reviews: reviews,
	}
}
