package service

import (
	"context"
	"errors"
	"testing"

	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/service/cart/model"
	"ecommerce-system/internal/service/cart/repository"
)

type mockCartRepo struct {
	getByUserIDFn     func(ctx context.Context, userID uint64) ([]*model.Cart, error)
	addItemFn         func(ctx context.Context, cart *model.Cart) error
	updateQuantityFn  func(ctx context.Context, userID, skuID uint64, quantity int) error
	removeItemFn      func(ctx context.Context, userID uint64, skuIDs []uint64) error
	clearCartFn       func(ctx context.Context, userID uint64) error
	selectItemFn      func(ctx context.Context, userID, skuID uint64, isSelected int8) error
	batchSelectFn     func(ctx context.Context, userID uint64, skuIDs []uint64, isSelected int8) error
	syncToDBFn        func(ctx context.Context, userID uint64) error
	updateCalledCount int
}

var _ repository.CartRepository = (*mockCartRepo)(nil)

func (m *mockCartRepo) GetByUserID(ctx context.Context, userID uint64) ([]*model.Cart, error) {
	if m.getByUserIDFn == nil {
		return nil, nil
	}
	return m.getByUserIDFn(ctx, userID)
}

func (m *mockCartRepo) AddItem(ctx context.Context, cart *model.Cart) error {
	if m.addItemFn == nil {
		return nil
	}
	return m.addItemFn(ctx, cart)
}

func (m *mockCartRepo) UpdateQuantity(ctx context.Context, userID, skuID uint64, quantity int) error {
	m.updateCalledCount++
	if m.updateQuantityFn == nil {
		return nil
	}
	return m.updateQuantityFn(ctx, userID, skuID, quantity)
}

func (m *mockCartRepo) RemoveItem(ctx context.Context, userID uint64, skuIDs []uint64) error {
	if m.removeItemFn == nil {
		return nil
	}
	return m.removeItemFn(ctx, userID, skuIDs)
}

func (m *mockCartRepo) ClearCart(ctx context.Context, userID uint64) error {
	if m.clearCartFn == nil {
		return nil
	}
	return m.clearCartFn(ctx, userID)
}

func (m *mockCartRepo) SelectItem(ctx context.Context, userID, skuID uint64, isSelected int8) error {
	if m.selectItemFn == nil {
		return nil
	}
	return m.selectItemFn(ctx, userID, skuID, isSelected)
}

func (m *mockCartRepo) BatchSelect(ctx context.Context, userID uint64, skuIDs []uint64, isSelected int8) error {
	if m.batchSelectFn == nil {
		return nil
	}
	return m.batchSelectFn(ctx, userID, skuIDs, isSelected)
}

func (m *mockCartRepo) SyncToDB(ctx context.Context, userID uint64) error {
	if m.syncToDBFn == nil {
		return nil
	}
	return m.syncToDBFn(ctx, userID)
}

func TestCartLogicAddItemSuccess(t *testing.T) {
	repo := &mockCartRepo{
		addItemFn: func(_ context.Context, cart *model.Cart) error {
			if cart.UserID != 7 || cart.SkuID != 1001 || cart.Quantity != 2 {
				t.Fatalf("unexpected cart payload: %+v", cart)
			}
			if cart.IsSelected != 1 {
				t.Fatalf("expected IsSelected=1, got %d", cart.IsSelected)
			}
			return nil
		},
	}

	logic := NewCartLogic(repo)
	resp, err := logic.AddItem(context.Background(), &AddItemRequest{
		UserID:   7,
		SkuID:    1001,
		Quantity: 2,
	})
	if err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}
	if resp == nil || resp.Cart == nil {
		t.Fatal("expected response cart, got nil")
	}
	if resp.Cart.IsSelected != 1 {
		t.Fatalf("expected response cart IsSelected=1, got %d", resp.Cart.IsSelected)
	}
}

func TestCartLogicGetCartWrapsRepoError(t *testing.T) {
	repo := &mockCartRepo{
		getByUserIDFn: func(_ context.Context, _ uint64) ([]*model.Cart, error) {
			return nil, errors.New("redis down")
		},
	}

	logic := NewCartLogic(repo)
	_, err := logic.GetCart(context.Background(), &GetCartRequest{UserID: 1})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	bizErr, ok := err.(*apperrors.BusinessError)
	if !ok {
		t.Fatalf("expected BusinessError, got %T", err)
	}
	if bizErr.Code != apperrors.CodeInternalError {
		t.Fatalf("expected CodeInternalError, got %d", bizErr.Code)
	}
}

func TestCartLogicUpdateQuantityValidateQuantity(t *testing.T) {
	repo := &mockCartRepo{}
	logic := NewCartLogic(repo)

	err := logic.UpdateQuantity(context.Background(), &UpdateQuantityRequest{
		UserID:   1,
		SkuID:    2,
		Quantity: 0,
	})
	if err == nil {
		t.Fatal("expected error for invalid quantity, got nil")
	}

	bizErr, ok := err.(*apperrors.BusinessError)
	if !ok {
		t.Fatalf("expected BusinessError, got %T", err)
	}
	if bizErr.Code != apperrors.CodeInvalidParam {
		t.Fatalf("expected CodeInvalidParam, got %d", bizErr.Code)
	}
	if repo.updateCalledCount != 0 {
		t.Fatalf("expected repo UpdateQuantity not called, called=%d", repo.updateCalledCount)
	}
}
