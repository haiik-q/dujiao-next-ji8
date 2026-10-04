package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dujiao-next/internal/modules/xcheck/domain"
)

func TestNormalizeHandle(t *testing.T) {
	cases := map[string]string{
		"ji8dotai":                        "ji8dotai",
		"  @ji8dotai ":                    "ji8dotai",
		"https://x.com/ji8dotai":          "ji8dotai",
		"x.com/ji8dotai?s=21":             "ji8dotai",
		"https://twitter.com/Ji8_Ai/":     "Ji8_Ai",
		"https://mobile.twitter.com/abc1": "abc1",
	}
	for in, want := range cases {
		got, err := NormalizeHandle(in)
		if err != nil || got != want {
			t.Errorf("NormalizeHandle(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "@", "has space", "abcdefghijklmnop", "中文名", "https://x.com/"} {
		if _, err := NormalizeHandle(bad); !errors.Is(err, domain.ErrInvalidHandle) {
			t.Errorf("NormalizeHandle(%q) err = %v; want ErrInvalidHandle", bad, err)
		}
	}
}

func boolPtr(v bool) *bool { return &v }

func TestEvaluate(t *testing.T) {
	now := time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name     string
		profile  domain.Profile
		eligible bool
		reason   string
	}{
		{"not found", domain.Profile{}, false, domain.ReasonNotFound},
		{"suspended", domain.Profile{Found: true, Suspended: true}, false, domain.ReasonSuspended},
		{"eligible per x", domain.Profile{Found: true, ScreenName: "a", GiftingEligible: boolPtr(true)}, true, ""},
		{"organization", domain.Profile{Found: true, IsBlueVerified: true, VerifiedType: "Business", GiftingEligible: boolPtr(false)}, false, domain.ReasonVerifiedOrganization},
		{"already premium", domain.Profile{Found: true, IsBlueVerified: true, GiftingEligible: boolPtr(false)}, false, domain.ReasonAlreadyPremium},
		{"protected", domain.Profile{Found: true, Protected: true, GiftingEligible: boolPtr(false)}, false, domain.ReasonProtected},
		{"hidden subscriptions", domain.Profile{Found: true, HiddenSubscriptions: true, GiftingEligible: boolPtr(false)}, false, domain.ReasonHiddenSubscriptions},
		{"x says no without known reason", domain.Profile{Found: true, GiftingEligible: boolPtr(false)}, false, domain.ReasonNotEligible},
		{"field missing, derive eligible", domain.Profile{Found: true}, true, ""},
		{"field missing, derive premium", domain.Profile{Found: true, IsBlueVerified: true}, false, domain.ReasonAlreadyPremium},
	}
	for _, tc := range cases {
		got := Evaluate("a", tc.profile, now)
		if got.Eligible != tc.eligible || got.Reason != tc.reason {
			t.Errorf("%s: got eligible=%v reason=%q; want %v %q", tc.name, got.Eligible, got.Reason, tc.eligible, tc.reason)
		}
	}
}

type countingLookup struct {
	calls   int
	profile domain.Profile
	err     error
}

func (l *countingLookup) Lookup(context.Context, string) (domain.Profile, error) {
	l.calls++
	return l.profile, l.err
}

func TestCheckCachesResults(t *testing.T) {
	lookup := &countingLookup{profile: domain.Profile{Found: true, ScreenName: "Abc", GiftingEligible: boolPtr(true)}}
	svc := NewService(lookup)
	now := time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	for _, in := range []string{"abc", "@ABC", "https://x.com/abc"} {
		res, err := svc.Check(context.Background(), in)
		if err != nil || !res.Eligible || res.Handle != "Abc" {
			t.Fatalf("Check(%q) = %+v, %v", in, res, err)
		}
	}
	if lookup.calls != 1 {
		t.Fatalf("lookup calls = %d; want 1 (cached)", lookup.calls)
	}
	now = now.Add(CacheTTL + time.Second)
	if _, err := svc.Check(context.Background(), "abc"); err != nil {
		t.Fatal(err)
	}
	if lookup.calls != 2 {
		t.Fatalf("lookup calls after expiry = %d; want 2", lookup.calls)
	}
}

func TestCheckDoesNotCacheErrors(t *testing.T) {
	lookup := &countingLookup{err: domain.ErrUnavailable}
	svc := NewService(lookup)
	for i := 0; i < 2; i++ {
		if _, err := svc.Check(context.Background(), "abc"); !errors.Is(err, domain.ErrUnavailable) {
			t.Fatalf("err = %v", err)
		}
	}
	if lookup.calls != 2 {
		t.Fatalf("lookup calls = %d; want 2", lookup.calls)
	}
}

func TestVerifyAlwaysAsksXAndRejectsIneligible(t *testing.T) {
	lookup := &countingLookup{profile: domain.Profile{Found: true, ScreenName: "abc", GiftingEligible: boolPtr(true)}}
	svc := NewService(lookup)
	if _, err := svc.Check(context.Background(), "abc"); err != nil {
		t.Fatal(err)
	}
	// 自检缓存里是「可以接收」，但下单前 X 改口了：必须以最新结果为准
	lookup.profile = domain.Profile{Found: true, ScreenName: "abc", GiftingEligible: boolPtr(false)}
	if err := svc.Verify(context.Background(), "@abc"); !errors.Is(err, domain.ErrNotEligible) {
		t.Fatalf("Verify err = %v; want ErrNotEligible", err)
	}
	if lookup.calls != 2 {
		t.Fatalf("lookup calls = %d; want 2 (Verify bypasses cache)", lookup.calls)
	}
	// 最新结果写回缓存，自检页也随之变为不符合
	if res, err := svc.Check(context.Background(), "abc"); err != nil || res.Eligible {
		t.Fatalf("Check after Verify = %+v, %v; want not eligible from cache", res, err)
	}
	if err := svc.Verify(context.Background(), "bad handle!"); !errors.Is(err, domain.ErrInvalidHandle) {
		t.Fatalf("Verify invalid handle err = %v", err)
	}
}

func TestVerifyFallsBackToRecentEligibleWhenXUnavailable(t *testing.T) {
	lookup := &countingLookup{profile: domain.Profile{Found: true, ScreenName: "abc", GiftingEligible: boolPtr(true)}}
	svc := NewService(lookup)
	now := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }
	if err := svc.Verify(context.Background(), "abc"); err != nil {
		t.Fatalf("Verify eligible err = %v", err)
	}
	lookup.err = errors.New("x 429")
	if err := svc.Verify(context.Background(), "abc"); err != nil {
		t.Fatalf("Verify with recent eligible result should pass, got %v", err)
	}
	if err := svc.Verify(context.Background(), "other"); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("Verify without cached result err = %v; want ErrUnavailable", err)
	}
	now = now.Add(CacheTTL + time.Second)
	if err := svc.Verify(context.Background(), "abc"); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("Verify after cache expiry err = %v; want ErrUnavailable", err)
	}
}
