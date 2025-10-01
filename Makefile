GOCMD=go

UNIT_TEST_FLAGS=-short

ROOT_DIR=.
PROTO_DIR=backend/proto
SHARED_DIR=backend/shared
API_GATEWAY_DIR=backend/api-gateway
AUTH_SERVICE_DIR=backend/auth-service
ESSAY_SERVICE_DIR=backend/essay-service
REVIEW_SERVICE_DIR=backend/review-service
NOTIFICATION_SERVICE_DIR=backend/notification-service

.PHONY: unit-test unit-test-shared unit-test-api-gateway unit-test-auth unit-test-essay unit-test-review unit-test-notification
unit-test: unit-test-shared unit-test-api-gateway unit-test-auth unit-test-essay unit-test-review unit-test-notification

unit-test-shared:
	cd $(SHARED_DIR) && go test ./... $(UNIT_TEST_FLAGS)

unit-test-api-gateway:
	cd $(API_GATEWAY_DIR) && go test ./... $(UNIT_TEST_FLAGS)

unit-test-auth:
	cd $(AUTH_SERVICE_DIR) && go test ./... $(UNIT_TEST_FLAGS)

unit-test-essay:
	cd $(ESSAY_SERVICE_DIR) && go test ./... $(UNIT_TEST_FLAGS)

unit-test-review:
	cd $(REVIEW_SERVICE_DIR) && go test ./... $(UNIT_TEST_FLAGS)

unit-test-notification:
	cd $(NOTIFICATION_SERVICE_DIR) && go test ./... $(UNIT_TEST_FLAGS)


.PHONY: mod-tidy mod-tidy-proto mod-tidy-shared mod-tidy-api-gateway mod-tidy-auth mod-tidy-essay mod-tidy-review mod-tidy-notification
mod-tidy: mod-tidy-proto mod-tidy-shared mod-tidy-api-gateway mod-tidy-auth mod-tidy-essay mod-tidy-review mod-tidy-notification

mod-tidy-proto:
	cd $(PROTO_DIR) && go mod tidy

mod-tidy-shared:
	cd $(SHARED_DIR) && go mod tidy

mod-tidy-api-gateway:
	cd $(API_GATEWAY_DIR) && go mod tidy

mod-tidy-auth:
	cd $(AUTH_SERVICE_DIR) && go mod tidy

mod-tidy-essay:
	cd $(ESSAY_SERVICE_DIR) && go mod tidy

mod-tidy-review:
	cd $(REVIEW_SERVICE_DIR) && go mod tidy

mod-tidy-notification:
	cd $(NOTIFICATION_SERVICE_DIR) && go mod tidy
