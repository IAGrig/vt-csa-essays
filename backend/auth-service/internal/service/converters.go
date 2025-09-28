package service

import (
	"github.com/IAGrig/vt-csa-essays/backend/auth-service/internal/models"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
)


func toProtoUserResponse(u models.User) *pb.UserResponse {
	return &pb.UserResponse{
		Id:        int32(u.ID),
		Username:  u.Username,
		CreatedAt: u.CreatedAt.Unix(),
	}
}
