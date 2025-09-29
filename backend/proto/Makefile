PROTOC := protoc

PROTO_DIR := .
NOTIFICATION_PROTO := $(PROTO_DIR)/notification/notification.proto
ESSAY_PROTO := $(PROTO_DIR)/essay/essay.proto
REVIEW_PROTO := $(PROTO_DIR)/review/review.proto
USER_PROTO := $(PROTO_DIR)/user/user.proto

OUT_DIR := $(PROTO_DIR)

all: generate-all

generate-all: $(NOTIFICATION_PROTO) $(ESSAY_PROTO) $(REVIEW_PROTO) $(USER_PROTO)
	@$(PROTOC) --go_out=$(OUT_DIR) --go_opt=paths=source_relative --go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative $(NOTIFICATION_PROTO)
	@$(PROTOC) --go_out=$(OUT_DIR) --go_opt=paths=source_relative --go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative $(ESSAY_PROTO)
	@$(PROTOC) --go_out=$(OUT_DIR) --go_opt=paths=source_relative --go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative $(REVIEW_PROTO)
	@$(PROTOC) --go_out=$(OUT_DIR) --go_opt=paths=source_relative --go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative $(USER_PROTO)

generate-essay: $(ESSAY_PROTO)
	@$(PROTOC) --go_out=$(OUT_DIR) --go_opt=paths=source_relative --go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative $(ESSAY_PROTO)

generate-notification: $(NOTIFICATION_PROTO)
	@$(PROTOC) --go_out=$(OUT_DIR) --go_opt=paths=source_relative --go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative $(NOTIFICATION_PROTO)

generate-review: $(REVIEW_PROTO)
	@$(PROTOC) --go_out=$(OUT_DIR) --go_opt=paths=source_relative --go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative $(REVIEW_PROTO)

generate-user: $(USER_PROTO)
	@$(PROTOC) --go_out=$(OUT_DIR) --go_opt=paths=source_relative --go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative $(USER_PROTO)


.PHONY: all generate-all generate-essay generate-notification generate-review generate-user
