package producthttp

import (
	"testing"

	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	memberleveldomain "github.com/dujiao-next/internal/modules/memberlevel/domain"
	"github.com/dujiao-next/internal/shared/money"
	"github.com/shopspring/decimal"
)

type stubMemberLevelPricing struct {
	prices []memberleveldomain.MemberLevelPrice
}

func (s stubMemberLevelPricing) GetLevelPricesByProduct(productID uint) ([]memberleveldomain.MemberLevelPrice, error) {
	return s.prices, nil
}

func (s stubMemberLevelPricing) ResolveMemberPrice(levelID, productID, skuID uint, basePrice decimal.Decimal) (decimal.Decimal, decimal.Decimal) {
	for _, p := range s.prices {
		if p.MemberLevelID == levelID && (p.SKUID == skuID || p.SKUID == 0) {
			return p.PriceAmount.Decimal, basePrice.Sub(p.PriceAmount.Decimal)
		}
	}
	return basePrice, decimal.Zero
}

// ji8：等级特价是给个别客户的批发价，公开接口只返回访客自己等级的价格。
func TestDecoratePublicProductOnlyExposesViewersOwnMemberPrices(t *testing.T) {
	h := &PublicHandler{memberLevels: stubMemberLevelPricing{prices: []memberleveldomain.MemberLevelPrice{
		{MemberLevelID: 1, ProductID: 1, PriceAmount: money.FromDecimal(decimal.NewFromInt(19))},
		{MemberLevelID: 2, ProductID: 1, PriceAmount: money.FromDecimal(decimal.NewFromInt(18))},
	}}}
	newProduct := func() *productdomain.Product {
		return &productdomain.Product{
			ID:          1,
			PriceAmount: money.FromDecimal(decimal.NewFromInt(20)),
			SKUs: []productdomain.ProductSKU{
				{ID: 11, ProductID: 1, IsActive: true, PriceAmount: money.FromDecimal(decimal.NewFromInt(20))},
			},
		}
	}

	guest, err := h.decoratePublicProduct(newProduct(), nil)
	if err != nil {
		t.Fatalf("decoratePublicProduct failed: %v", err)
	}
	if len(guest.MemberPrices) != 0 {
		t.Fatalf("guest should see no member prices, got: %+v", guest.MemberPrices)
	}

	member, err := h.decoratePublicProduct(newProduct(), nil, 1)
	if err != nil {
		t.Fatalf("decoratePublicProduct failed: %v", err)
	}
	if len(member.MemberPrices) != 1 || member.MemberPrices[0].MemberLevelID != 1 ||
		!member.MemberPrices[0].PriceAmount.Decimal.Equal(decimal.NewFromInt(19)) {
		t.Fatalf("level 1 should only see its own price 19, got: %+v", member.MemberPrices)
	}
}
