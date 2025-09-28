package converters

import (
	"github.com/gin-gonic/gin"

	pb "github.com/IAGrig/vt-csa-essays/backend/proto/user"
)

func MarshalProtoUserResponse(u *pb.UserResponse) gin.H {
	if u == nil {
		return gin.H{}
	}
	return gin.H{
		"id":         u.Id,
		"username":   u.Username,
		"created_at": u.CreatedAt,
	}
}
