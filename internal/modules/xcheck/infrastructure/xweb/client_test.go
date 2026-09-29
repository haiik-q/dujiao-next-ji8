package xweb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseOperation(t *testing.T) {
	script := `x,360985(e){e.exports={queryId:"KybxDj9RrADIITXlGG8kpw",operationName:"UserByScreenName",operationType:"query",metadata:{featureSwitches:["hidden_profile_subscriptions_enabled","subscriptions_feature_can_gift_premium"],fieldToggles:["withPayments","withAuxiliaryUserLabels"]}}},y="Bearer AAAAAAAAAAAAAAAAAAAAAFakeToken%3Dabc"`
	op, err := parseOperation(script)
	if err != nil {
		t.Fatal(err)
	}
	if op.queryID != "KybxDj9RrADIITXlGG8kpw" || op.bearer != "AAAAAAAAAAAAAAAAAAAAAFakeToken=abc" {
		t.Fatalf("op = %+v", op)
	}
	if len(op.features) != 2 || op.features[1] != "subscriptions_feature_can_gift_premium" || len(op.fieldToggles) != 2 {
		t.Fatalf("features = %v toggles = %v", op.features, op.fieldToggles)
	}
	if _, err := parseOperation("no operation here"); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseProfile(t *testing.T) {
	p, err := parseProfile([]byte(`{"data":{"user":{"result":{"__typename":"User","core":{"screen_name":"XDevelopers","name":"Developers"},"avatar":{"image_url":"https://pbs.twimg.com/a.jpg"},"is_blue_verified":true,"verification":{"verified_type":"Business"},"privacy":{"protected":false},"has_hidden_subscriptions_on_profile":false,"premium_gifting_eligible":false}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if !p.Found || p.ScreenName != "XDevelopers" || p.VerifiedType != "Business" || !p.IsBlueVerified || p.GiftingEligible == nil || *p.GiftingEligible {
		t.Fatalf("profile = %+v", p)
	}

	p, err = parseProfile([]byte(`{"data":{}}`))
	if err != nil || p.Found {
		t.Fatalf("not found: %+v %v", p, err)
	}

	p, err = parseProfile([]byte(`{"data":{"user":{"result":{"__typename":"UserUnavailable","reason":"Suspended"}}}}`))
	if err != nil || !p.Found || !p.Suspended {
		t.Fatalf("suspended: %+v %v", p, err)
	}

	if _, err := parseProfile([]byte(`{"errors":[{"message":"boom"}]}`)); err == nil {
		t.Fatal("expected error for graphql errors")
	}
}

func TestLoadAccounts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x-accounts.txt")
	content := "# comment\nuser1|pw|mail|mailpw|refresh|client|2fa|ct0value|authvalue\n\nbroken|line\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	accounts, err := LoadAccounts(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 || accounts[0].Name != "user1" || accounts[0].CT0 != "ct0value" || accounts[0].AuthToken != "authvalue" {
		t.Fatalf("accounts = %+v", accounts)
	}
	missing, err := LoadAccounts(filepath.Join(dir, "missing.txt"))
	if err != nil || missing != nil {
		t.Fatalf("missing file: %v %v", missing, err)
	}
}

func TestPickAccountSkipsCooling(t *testing.T) {
	c := NewClient([]Account{{Name: "a", CT0: "1", AuthToken: "1"}, {Name: "b", CT0: "2", AuthToken: "2"}})
	first := c.pickAccount()
	second := c.pickAccount()
	if first.Name == second.Name {
		t.Fatalf("round robin picked %s twice", first.Name)
	}
	c.accounts[0].coolUntil = c.now().Add(1e12)
	for i := 0; i < 3; i++ {
		if got := c.pickAccount(); got.Name != "b" {
			t.Fatalf("picked %s while a is cooling", got.Name)
		}
	}
	c.accounts[1].coolUntil = c.now().Add(1e12)
	if got := c.pickAccount(); got != nil {
		t.Fatalf("picked %s while all cooling", got.Name)
	}
}
