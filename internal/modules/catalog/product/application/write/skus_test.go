package productwrite

import (
	"errors"
	"sort"
	"strings"
	"testing"

	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"

	"github.com/dujiao-next/internal/constants"
	productcontract "github.com/dujiao-next/internal/modules/catalog/product/contract"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/shopspring/decimal"
)

func TestSyncSingleProductSKUMultipleRowsKeepsSingleActive(t *testing.T) {
	service := NewWriteService(Options{})
	repo := newSyncSingleSKURepo(t)
	productID := uint(2001)

	inactiveDefault := productdomain.ProductSKU{
		ProductID:        productID,
		SKUCode:          productdomain.DefaultSKUCode,
		PriceAmount:      money.FromDecimal(decimal.NewFromInt(10)),
		ManualStockTotal: 9,
		IsActive:         false,
		SortOrder:        0,
	}
	firstActive := productdomain.ProductSKU{
		ProductID:        productID,
		SKUCode:          "A",
		PriceAmount:      money.FromDecimal(decimal.NewFromInt(20)),
		ManualStockTotal: 2,
		IsActive:         true,
		SortOrder:        2,
	}
	secondActive := productdomain.ProductSKU{
		ProductID:        productID,
		SKUCode:          "B",
		PriceAmount:      money.FromDecimal(decimal.NewFromInt(30)),
		ManualStockTotal: 4,
		IsActive:         true,
		SortOrder:        1,
	}
	if err := repo.Create(&inactiveDefault); err != nil {
		t.Fatalf("create inactive default sku failed: %v", err)
	}
	inactiveDefault.IsActive = false
	if err := repo.Update(&inactiveDefault); err != nil {
		t.Fatalf("update inactive default sku failed: %v", err)
	}
	if err := repo.Create(&firstActive); err != nil {
		t.Fatalf("create first active sku failed: %v", err)
	}
	if err := repo.Create(&secondActive); err != nil {
		t.Fatalf("create second active sku failed: %v", err)
	}

	targetPrice := decimal.RequireFromString("88.88")
	if err := service.syncSingleProductSKU(repo, nil, productID, "", targetPrice, decimal.Zero, 5); err != nil {
		t.Fatalf("sync single sku failed: %v", err)
	}

	skus, err := repo.ListByProduct(productID, false)
	if err != nil {
		t.Fatalf("list sku failed: %v", err)
	}
	if len(skus) != 1 {
		t.Fatalf("expected exactly one sku row after collapsing to single spec, got %d: %+v", len(skus), skus)
	}
	// 存续的必须是 DEFAULT 行：单规格模式的唯一表示，且承载价格与库存。
	if skus[0].ID != inactiveDefault.ID {
		t.Fatalf("expected surviving row to be the DEFAULT sku id=%d, got id=%d", inactiveDefault.ID, skus[0].ID)
	}
	if !strings.EqualFold(skus[0].SKUCode, productdomain.DefaultSKUCode) {
		t.Fatalf("expected surviving row code DEFAULT, got %q", skus[0].SKUCode)
	}
	if !skus[0].IsActive {
		t.Fatalf("expected surviving DEFAULT row to be active: %+v", skus[0])
	}
	if !skus[0].PriceAmount.Equal(targetPrice) || skus[0].ManualStockTotal != 5 {
		t.Fatalf("unexpected synchronized SKU: %+v", skus[0])
	}
}

func TestSyncSingleProductSKURenamesSurvivorToDefault(t *testing.T) {
	service := NewWriteService(Options{})
	repo := newSyncSingleSKURepo(t)
	productID := uint(2003)

	// 没有任何 DEFAULT 行时，退化挑一行现存规格改名，保证存续行编码为 DEFAULT。
	first := productdomain.ProductSKU{
		ProductID:        productID,
		SKUCode:          "A",
		PriceAmount:      money.FromDecimal(decimal.NewFromInt(20)),
		ManualStockTotal: 2,
		IsActive:         true,
		SortOrder:        2,
	}
	second := productdomain.ProductSKU{
		ProductID:        productID,
		SKUCode:          "B",
		PriceAmount:      money.FromDecimal(decimal.NewFromInt(30)),
		ManualStockTotal: 4,
		IsActive:         true,
		SortOrder:        1,
	}
	if err := repo.Create(&first); err != nil {
		t.Fatalf("create sku A failed: %v", err)
	}
	if err := repo.Create(&second); err != nil {
		t.Fatalf("create sku B failed: %v", err)
	}

	targetPrice := decimal.RequireFromString("12.34")
	if err := service.syncSingleProductSKU(repo, nil, productID, "", targetPrice, decimal.Zero, 7); err != nil {
		t.Fatalf("sync single sku failed: %v", err)
	}

	skus, err := repo.ListByProduct(productID, false)
	if err != nil {
		t.Fatalf("list sku failed: %v", err)
	}
	if len(skus) != 1 {
		t.Fatalf("expected exactly one sku row, got %d: %+v", len(skus), skus)
	}
	if !strings.EqualFold(skus[0].SKUCode, productdomain.DefaultSKUCode) {
		t.Fatalf("expected survivor renamed to DEFAULT, got %q", skus[0].SKUCode)
	}
	if !skus[0].PriceAmount.Equal(targetPrice) || skus[0].ManualStockTotal != 7 {
		t.Fatalf("unexpected synchronized SKU: %+v", skus[0])
	}
}

// 自动发货商品通过单规格路径删规格时，同样不能删掉还压着卡密库存的行。
func TestSyncSingleProductSKURejectsRemovingSKUWithCardSecretStock(t *testing.T) {
	service := NewWriteService(Options{})
	repo := newSyncSingleSKURepo(t)
	productID := uint(2004)

	target := productdomain.ProductSKU{
		ProductID:        productID,
		SKUCode:          productdomain.DefaultSKUCode,
		PriceAmount:      money.FromDecimal(decimal.NewFromInt(10)),
		ManualStockTotal: 0,
		IsActive:         true,
		SortOrder:        0,
	}
	withStock := productdomain.ProductSKU{
		ProductID:        productID,
		SKUCode:          "A",
		PriceAmount:      money.FromDecimal(decimal.NewFromInt(20)),
		ManualStockTotal: 0,
		IsActive:         true,
		SortOrder:        1,
	}
	if err := repo.Create(&target); err != nil {
		t.Fatalf("create default sku failed: %v", err)
	}
	if err := repo.Create(&withStock); err != nil {
		t.Fatalf("create sku A failed: %v", err)
	}

	cardSecrets := &memoryCardSecretRepo{available: map[uint]int64{withStock.ID: 3}}
	err := service.syncSingleProductSKU(repo, cardSecrets, productID, constants.FulfillmentTypeAuto, decimal.NewFromInt(10), decimal.Zero, 0)
	if !errors.Is(err, productcontract.ErrProductSKUHasCardSecretStock) {
		t.Fatalf("expected ErrProductSKUHasCardSecretStock, got %v", err)
	}

	// 拒绝时不应留下半删状态。
	skus, listErr := repo.ListByProduct(productID, false)
	if listErr != nil {
		t.Fatalf("list sku failed: %v", listErr)
	}
	if len(skus) != 2 {
		t.Fatalf("expected both rows to survive rejection, got %d: %+v", len(skus), skus)
	}
}

type memoryCardSecretRepo struct {
	available map[uint]int64
}

func (repo *memoryCardSecretRepo) CountByProduct(_ uint, skuID uint) (int64, int64, int64, error) {
	available := repo.available[skuID]
	return available, available, 0, nil
}

func TestSyncSingleProductSKUNoActivePrefersDefaultCode(t *testing.T) {
	service := NewWriteService(Options{})
	repo := newSyncSingleSKURepo(t)
	productID := uint(2002)

	inactiveA := productdomain.ProductSKU{
		ProductID:        productID,
		SKUCode:          "A",
		PriceAmount:      money.FromDecimal(decimal.NewFromInt(10)),
		ManualStockTotal: 3,
		IsActive:         false,
		SortOrder:        1,
	}
	inactiveDefault := productdomain.ProductSKU{
		ProductID:        productID,
		SKUCode:          productdomain.DefaultSKUCode,
		PriceAmount:      money.FromDecimal(decimal.NewFromInt(20)),
		ManualStockTotal: 8,
		IsActive:         false,
		SortOrder:        0,
	}
	if err := repo.Create(&inactiveA); err != nil {
		t.Fatalf("create inactive sku A failed: %v", err)
	}
	inactiveA.IsActive = false
	if err := repo.Update(&inactiveA); err != nil {
		t.Fatalf("update inactive sku A failed: %v", err)
	}
	if err := repo.Create(&inactiveDefault); err != nil {
		t.Fatalf("create inactive default sku failed: %v", err)
	}
	inactiveDefault.IsActive = false
	if err := repo.Update(&inactiveDefault); err != nil {
		t.Fatalf("update inactive default sku failed: %v", err)
	}

	targetPrice := decimal.RequireFromString("19.90")
	if err := service.syncSingleProductSKU(repo, nil, productID, "", targetPrice, decimal.Zero, 6); err != nil {
		t.Fatalf("sync single sku failed: %v", err)
	}

	skus, err := repo.ListByProduct(productID, false)
	if err != nil {
		t.Fatalf("list sku failed: %v", err)
	}
	activeCount := 0
	for _, sku := range skus {
		if !sku.IsActive {
			continue
		}
		activeCount++
		if sku.ID != inactiveDefault.ID {
			t.Fatalf("expected default sku id=%d to be active, got id=%d", inactiveDefault.ID, sku.ID)
		}
		if !sku.PriceAmount.Equal(targetPrice) || sku.ManualStockTotal != 6 {
			t.Fatalf("unexpected synchronized DEFAULT SKU: %+v", sku)
		}
	}
	if activeCount != 1 {
		t.Fatalf("expected exactly one active sku, got %d", activeCount)
	}
}

type memorySKURepository struct {
	rows   []productdomain.ProductSKU
	nextID uint
}

func (repo *memorySKURepository) ListByProduct(productID uint, onlyActive bool) ([]productdomain.ProductSKU, error) {
	rows := make([]productdomain.ProductSKU, 0, len(repo.rows))
	for _, row := range repo.rows {
		if row.ProductID == productID && (!onlyActive || row.IsActive) {
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].SortOrder != rows[j].SortOrder {
			return rows[i].SortOrder > rows[j].SortOrder
		}
		return rows[i].ID < rows[j].ID
	})
	return rows, nil
}

func (repo *memorySKURepository) Create(item *productdomain.ProductSKU) error {
	if item.ID == 0 {
		repo.nextID++
		item.ID = repo.nextID
	}
	repo.rows = append(repo.rows, *item)
	return nil
}

func (repo *memorySKURepository) Update(item *productdomain.ProductSKU) error {
	for index := range repo.rows {
		if repo.rows[index].ID == item.ID {
			repo.rows[index] = *item
			return nil
		}
	}
	return nil
}

func (repo *memorySKURepository) Delete(id uint) error {
	for index := range repo.rows {
		if repo.rows[index].ID == id {
			repo.rows = append(repo.rows[:index], repo.rows[index+1:]...)
			break
		}
	}
	return nil
}

func (repo *memorySKURepository) PurgeSoftDeletedByProductAndCode(uint, string) error {
	return nil
}

func newSyncSingleSKURepo(t *testing.T) SKURepository {
	t.Helper()
	return &memorySKURepository{}
}
