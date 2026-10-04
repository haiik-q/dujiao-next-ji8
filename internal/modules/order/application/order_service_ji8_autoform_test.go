package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	categorydomain "github.com/dujiao-next/internal/modules/catalog/category/domain"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	productgormstore "github.com/dujiao-next/internal/modules/catalog/product/store/gormstore"
	promotiondomain "github.com/dujiao-next/internal/modules/promotion/domain"
	promotiongormstore "github.com/dujiao-next/internal/modules/promotion/infrastructure/gormstore"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ji8：自动发货（卡密库存）商品配置了下单表单时，下单必须按表单校验并保存提交值（X 会员的 x_handle）。
func TestBuildOrderResultValidatesManualFormForAutoProduct(t *testing.T) {
	dsn := fmt.Sprintf("file:order_service_ji8_auto_form_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&categorydomain.Category{}, &productdomain.Product{}, &productdomain.ProductSKU{}, &promotiondomain.Promotion{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	now := time.Now()
	category := categorydomain.Category{Slug: "ji8-auto-form", NameJSON: jsonmap.JSON{"zh-CN": "X"}, CreatedAt: now}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category failed: %v", err)
	}
	product := productdomain.Product{
		CategoryID:      category.ID,
		Slug:            "ji8-auto-form-product",
		TitleJSON:       jsonmap.JSON{"zh-CN": "X Premium 3个月"},
		PriceAmount:     money.FromDecimal(decimal.NewFromInt(20)),
		PurchaseType:    constants.ProductPurchaseGuest,
		FulfillmentType: constants.FulfillmentTypeAuto,
		ManualFormSchemaJSON: jsonmap.JSON{
			"fields": []interface{}{
				map[string]interface{}{
					"key":      "x_handle",
					"type":     "text",
					"required": true,
					"regex":    "^@?[A-Za-z0-9_]{1,15}$",
					"label":    map[string]interface{}{"zh-CN": "X 用户名"},
				},
			},
		},
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}
	sku := productdomain.ProductSKU{
		ProductID:   product.ID,
		SKUCode:     productdomain.DefaultSKUCode,
		PriceAmount: money.FromDecimal(decimal.NewFromInt(20)),
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&sku).Error; err != nil {
		t.Fatalf("create sku failed: %v", err)
	}
	svc := NewOrderService(OrderServiceOptions{
		ProductStore:    productgormstore.NewProductStore(db),
		ProductSKUStore: productgormstore.NewSKUStore(db),
		PromotionRepo:   promotiongormstore.New(db),
		ExpireMinutes:   15,
	})
	build := func(form map[string]jsonmap.JSON) (*orderBuildResult, error) {
		return svc.buildOrderResult(orderCreateParams{
			UserID:         1,
			Items:          []CreateOrderItem{{ProductID: product.ID, SKUID: sku.ID, Quantity: 1}},
			ManualFormData: form,
		})
	}
	key := fmt.Sprint(product.ID)

	if _, err := build(nil); err == nil {
		t.Fatalf("expected missing x_handle to be rejected for auto product with form schema")
	}
	if _, err := build(map[string]jsonmap.JSON{key: {"x_handle": "bad handle!"}}); err == nil {
		t.Fatalf("expected invalid x_handle to be rejected")
	}
	res, err := build(map[string]jsonmap.JSON{key: {"x_handle": "@tasha13mj2"}})
	if err != nil {
		t.Fatalf("build order failed: %v", err)
	}
	if got := res.OrderItems[0].ManualFormSubmissionJSON["x_handle"]; got == nil || fmt.Sprint(got) == "" {
		t.Fatalf("expected x_handle saved on order item, got: %#v", res.OrderItems[0].ManualFormSubmissionJSON)
	}
}

type fakeXHandleVerifier struct {
	calls []string
	err   map[string]error
}

func (f *fakeXHandleVerifier) Verify(_ context.Context, handle string) error {
	f.calls = append(f.calls, handle)
	return f.err[handle]
}

// ji8：x_handle 不能接收赠送时创建订单直接失败（不锁库存、不发卡密）；预览不查 X。
func TestBuildOrderResultRejectsIneligibleXHandle(t *testing.T) {
	dsn := fmt.Sprintf("file:order_service_ji8_xverify_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&categorydomain.Category{}, &productdomain.Product{}, &productdomain.ProductSKU{}, &promotiondomain.Promotion{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	now := time.Now()
	category := categorydomain.Category{Slug: "ji8-xverify", NameJSON: jsonmap.JSON{"zh-CN": "X"}, CreatedAt: now}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category failed: %v", err)
	}
	product := productdomain.Product{
		CategoryID:      category.ID,
		Slug:            "ji8-xverify-product",
		TitleJSON:       jsonmap.JSON{"zh-CN": "X Premium 3个月"},
		PriceAmount:     money.FromDecimal(decimal.NewFromInt(23)),
		PurchaseType:    constants.ProductPurchaseGuest,
		FulfillmentType: constants.FulfillmentTypeAuto,
		ManualFormSchemaJSON: jsonmap.JSON{
			"fields": []interface{}{
				map[string]interface{}{"key": "x_handle", "type": "text", "required": true, "regex": "^@?[A-Za-z0-9_]{1,15}$"},
			},
		},
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product failed: %v", err)
	}
	sku := productdomain.ProductSKU{ProductID: product.ID, SKUCode: productdomain.DefaultSKUCode, PriceAmount: money.FromDecimal(decimal.NewFromInt(23)), IsActive: true, CreatedAt: now, UpdatedAt: now}
	if err := db.Create(&sku).Error; err != nil {
		t.Fatalf("create sku failed: %v", err)
	}
	errNotEligible := errors.New("not eligible")
	verifier := &fakeXHandleVerifier{err: map[string]error{"@jiuge105": errNotEligible}}
	svc := NewOrderService(OrderServiceOptions{
		ProductStore:    productgormstore.NewProductStore(db),
		ProductSKUStore: productgormstore.NewSKUStore(db),
		PromotionRepo:   promotiongormstore.New(db),
		XHandleVerifier: verifier,
		ExpireMinutes:   15,
	})
	key := fmt.Sprint(product.ID)
	build := func(handle string, skipForm bool) error {
		_, err := svc.buildOrderResult(orderCreateParams{
			UserID:              1,
			Items:               []CreateOrderItem{{ProductID: product.ID, SKUID: sku.ID, Quantity: 1}},
			ManualFormData:      map[string]jsonmap.JSON{key: {"x_handle": handle}},
			SkipManualFormCheck: skipForm,
		})
		return err
	}

	if err := build("@jiuge105", false); !errors.Is(err, errNotEligible) {
		t.Fatalf("expected ineligible x_handle to be rejected, got %v", err)
	}
	if err := build("tasha13mj2", false); err != nil {
		t.Fatalf("eligible x_handle rejected: %v", err)
	}
	if err := build("@jiuge105", true); err != nil {
		t.Fatalf("preview (skip form check) should not verify, got %v", err)
	}
	if len(verifier.calls) != 2 {
		t.Fatalf("verifier calls = %v; want 2 (preview must not query X)", verifier.calls)
	}
}
