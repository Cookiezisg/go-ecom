package cart

import (
	"strings"
	"testing"
	"time"

	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/service/cart/model"
	"ecommerce-system/internal/service/cart/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConvertErrorBusinessErrorToGrpcCode(t *testing.T) {
	err := convertError(apperrors.NewInvalidParamError("quantity invalid"))
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error, got %T", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", st.Code())
	}
	if !strings.Contains(st.Message(), "quantity invalid") {
		t.Fatalf("expected message to include original text, got %q", st.Message())
	}
}

func TestConvertErrorNil(t *testing.T) {
	if got := convertError(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestConvertDetailToProto(t *testing.T) {
	createdAt := time.Date(2026, 3, 4, 10, 20, 30, 0, time.UTC)
	updatedAt := time.Date(2026, 3, 4, 11, 20, 30, 0, time.UTC)
	cart := &model.Cart{
		ID:         1,
		UserID:     100,
		SkuID:      200,
		Quantity:   3,
		IsSelected: 1,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
	detail := &service.CartItemDetail{
		Cart:        cart,
		ProductID:   10,
		ProductName: "Test Product",
		SkuName:     "Red M",
		Price:       99.9,
		StockStatus: "in_stock",
	}

	got := convertDetailToProto(detail)
	if got == nil {
		t.Fatal("expected non-nil proto item")
	}
	if got.Id != 1 || got.UserId != 100 || got.SkuId != 200 {
		t.Fatalf("unexpected id mapping: %+v", got)
	}
	if got.Quantity != 3 || got.IsSelected != 1 {
		t.Fatalf("unexpected quantity/select mapping: %+v", got)
	}
	if got.CreatedAt != createdAt.Format(time.RFC3339) {
		t.Fatalf("unexpected created_at: %s", got.CreatedAt)
	}
	if got.ProductId != 10 || got.ProductName != "Test Product" {
		t.Fatalf("unexpected product info: %+v", got)
	}
	if got.StockStatus != "in_stock" {
		t.Fatalf("unexpected stock_status: %s", got.StockStatus)
	}
}

func TestConvertDetailToProtoNil(t *testing.T) {
	if got := convertDetailToProto(nil); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}
