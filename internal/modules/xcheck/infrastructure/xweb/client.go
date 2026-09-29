// Package xweb 通过 X 网页版内部 GraphQL（UserByScreenName）查询用户公开资料。
//
// 使用若干个登录小号的网页 Cookie（auth_token + ct0）轮询查询：每个账号每 15 分钟约 150 次额度，
// 被限流或失效时暂停该账号并换下一个。queryId、Bearer 与 features 随 X 前端版本变化，
// 首次使用和接口失效时从登录后的 x.com 主脚本里重新解析，不写死在代码里。
package xweb

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/modules/xcheck/domain"
)

const (
	userAgent         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0 Safari/537.36"
	homeURL           = "https://x.com/home"
	graphqlURL        = "https://x.com/i/api/graphql/%s/UserByScreenName"
	rediscoverAfter   = 5 * time.Minute
	authFailCooldown  = 30 * time.Minute
	rateLimitCooldown = 15 * time.Minute
	maxBodyBytes      = 8 << 20
)

var (
	mainScriptPattern = regexp.MustCompile(`https://abs\.twimg\.com/responsive-web/client-web[^"']*/main\.[a-z0-9]+\.js`)
	operationPattern  = regexp.MustCompile(`queryId:"([^"]+)",operationName:"UserByScreenName",operationType:"query",metadata:\{featureSwitches:\[([^\]]*)\],fieldToggles:\[([^\]]*)\]`)
	bearerPattern     = regexp.MustCompile(`AAAAAAAAAAAAAAAAAAAAA[A-Za-z0-9%]+`)
	quotedPattern     = regexp.MustCompile(`"([^"]+)"`)
)

// Account 是一个用于查询的 X 登录会话。
type Account struct {
	Name      string
	CT0       string
	AuthToken string
}

// LoadAccounts 读取账号文件：每行 "用户名|密码|邮箱|邮箱密码|refresh_token|client_id|2FA|ct0|auth_token"，
// 只使用第 1、8、9 列；# 开头的行和空行忽略。文件不存在时返回空列表（功能视为未启用）。
func LoadAccounts(path string) ([]Account, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var accounts []Account
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cols := strings.Split(line, "|")
		if len(cols) < 9 {
			continue
		}
		ct0, auth := strings.TrimSpace(cols[7]), strings.TrimSpace(cols[8])
		if ct0 == "" || auth == "" {
			continue
		}
		accounts = append(accounts, Account{Name: strings.TrimSpace(cols[0]), CT0: ct0, AuthToken: auth})
	}
	return accounts, scanner.Err()
}

type accountState struct {
	Account
	coolUntil time.Time
}

type operation struct {
	queryID      string
	bearer       string
	features     []string
	fieldToggles []string
}

// Client 查询 X 用户资料，并发安全。
type Client struct {
	httpClient *http.Client
	now        func() time.Time

	mu           sync.Mutex
	accounts     []*accountState
	next         int
	op           *operation
	lastDiscover time.Time
}

// NewClient 创建客户端；accounts 为空时所有查询返回 domain.ErrUnavailable。
func NewClient(accounts []Account) *Client {
	states := make([]*accountState, 0, len(accounts))
	for _, a := range accounts {
		states = append(states, &accountState{Account: a})
	}
	return &Client{
		httpClient: &http.Client{Timeout: 20 * time.Second},
		now:        time.Now,
		accounts:   states,
	}
}

// Lookup 按用户名查询资料。账号轮询：被限流/失效的账号暂停后换下一个，全部不可用时返回 ErrUnavailable。
func (c *Client) Lookup(ctx context.Context, handle string) (domain.Profile, error) {
	if c == nil || len(c.accounts) == 0 {
		return domain.Profile{}, domain.ErrUnavailable
	}
	rediscovered := false
	for attempt := 0; attempt < len(c.accounts)+1; attempt++ {
		acc := c.pickAccount()
		if acc == nil {
			return domain.Profile{}, domain.ErrUnavailable
		}
		op, err := c.operation(ctx, acc, false)
		if err != nil {
			logger.Warnw("xcheck_discover_failed", "account", acc.Name, "error", err)
			return domain.Profile{}, domain.ErrUnavailable
		}
		status, header, body, err := c.query(ctx, acc, op, handle)
		if err != nil {
			return domain.Profile{}, fmt.Errorf("%w: %v", domain.ErrUnavailable, err)
		}
		switch {
		case status == http.StatusOK:
			c.noteRateLimit(acc, header)
			return parseProfile(body)
		case status == http.StatusTooManyRequests:
			c.cool(acc, resetTime(header, c.now().Add(rateLimitCooldown)), "rate_limited")
		case status == http.StatusUnauthorized || status == http.StatusForbidden:
			c.cool(acc, c.now().Add(authFailCooldown), fmt.Sprintf("http_%d", status))
		case (status == http.StatusNotFound || status == http.StatusBadRequest) && !rediscovered:
			// queryId 或 features 过期：重新解析 X 前端后重试一次
			rediscovered = true
			if _, err := c.operation(ctx, acc, true); err != nil {
				logger.Warnw("xcheck_rediscover_failed", "account", acc.Name, "error", err)
				return domain.Profile{}, domain.ErrUnavailable
			}
		default:
			return domain.Profile{}, fmt.Errorf("%w: x responded %d", domain.ErrUnavailable, status)
		}
	}
	return domain.Profile{}, domain.ErrUnavailable
}

func (c *Client) pickAccount() *accountState {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	for i := 0; i < len(c.accounts); i++ {
		acc := c.accounts[(c.next+i)%len(c.accounts)]
		if now.Before(acc.coolUntil) {
			continue
		}
		c.next = (c.next + i + 1) % len(c.accounts)
		return acc
	}
	return nil
}

func (c *Client) cool(acc *accountState, until time.Time, why string) {
	c.mu.Lock()
	acc.coolUntil = until
	c.mu.Unlock()
	logger.Warnw("xcheck_account_paused", "account", acc.Name, "reason", why, "until", until)
}

// noteRateLimit 在额度用尽时提前暂停账号，避免下一次请求吃 429。
func (c *Client) noteRateLimit(acc *accountState, header http.Header) {
	remaining, err := strconv.Atoi(header.Get("x-rate-limit-remaining"))
	if err != nil || remaining > 0 {
		return
	}
	c.cool(acc, resetTime(header, c.now().Add(rateLimitCooldown)), "quota_exhausted")
}

func resetTime(header http.Header, fallback time.Time) time.Time {
	if sec, err := strconv.ParseInt(header.Get("x-rate-limit-reset"), 10, 64); err == nil && sec > 0 {
		return time.Unix(sec, 0)
	}
	return fallback
}

func (c *Client) operation(ctx context.Context, acc *accountState, force bool) (*operation, error) {
	c.mu.Lock()
	op, last := c.op, c.lastDiscover
	c.mu.Unlock()
	if op != nil && (!force || c.now().Sub(last) < rediscoverAfter) {
		return op, nil
	}
	discovered, err := c.discover(ctx, acc)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastDiscover = c.now()
	if err != nil {
		if c.op != nil {
			return c.op, nil
		}
		return nil, err
	}
	c.op = discovered
	logger.Infow("xcheck_operation_discovered", "query_id", discovered.queryID, "features", len(discovered.features))
	return discovered, nil
}

// discover 打开登录后的 x.com，找到主脚本并解析 UserByScreenName 的 queryId、features 与 Bearer。
func (c *Client) discover(ctx context.Context, acc *accountState) (*operation, error) {
	home, err := c.get(ctx, homeURL, acc)
	if err != nil {
		return nil, err
	}
	scriptURL := mainScriptPattern.FindString(home)
	if scriptURL == "" {
		return nil, errors.New("main script not found on x.com/home")
	}
	script, err := c.get(ctx, scriptURL, nil)
	if err != nil {
		return nil, err
	}
	return parseOperation(script)
}

func parseOperation(script string) (*operation, error) {
	m := operationPattern.FindStringSubmatch(script)
	if m == nil {
		return nil, errors.New("UserByScreenName operation not found")
	}
	rawBearer := bearerPattern.FindString(script)
	if rawBearer == "" {
		return nil, errors.New("bearer token not found")
	}
	bearer, err := url.QueryUnescape(rawBearer)
	if err != nil {
		return nil, err
	}
	return &operation{queryID: m[1], bearer: bearer, features: quotedList(m[2]), fieldToggles: quotedList(m[3])}, nil
}

func quotedList(raw string) []string {
	var out []string
	for _, m := range quotedPattern.FindAllStringSubmatch(raw, -1) {
		out = append(out, m[1])
	}
	return out
}

func (c *Client) get(ctx context.Context, target string, acc *accountState) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	if acc != nil {
		req.Header.Set("Cookie", "auth_token="+acc.AuthToken+"; ct0="+acc.CT0)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %d", target, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	return string(body), err
}

func (c *Client) query(ctx context.Context, acc *accountState, op *operation, handle string) (int, http.Header, []byte, error) {
	flags := func(names []string) string {
		m := make(map[string]bool, len(names))
		for _, n := range names {
			m[n] = true
		}
		b, _ := json.Marshal(m)
		return string(b)
	}
	variables, _ := json.Marshal(map[string]string{"screen_name": handle})
	params := url.Values{}
	params.Set("variables", string(variables))
	params.Set("features", flags(op.features))
	params.Set("fieldToggles", flags(op.fieldToggles))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(graphqlURL, op.queryID)+"?"+params.Encode(), nil)
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+op.bearer)
	req.Header.Set("x-csrf-token", acc.CT0)
	req.Header.Set("Cookie", "auth_token="+acc.AuthToken+"; ct0="+acc.CT0)
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	return resp.StatusCode, resp.Header, body, err
}

type userResult struct {
	Typename string `json:"__typename"`
	Reason   string `json:"reason"`
	Core     struct {
		ScreenName string `json:"screen_name"`
		Name       string `json:"name"`
	} `json:"core"`
	Avatar struct {
		ImageURL string `json:"image_url"`
	} `json:"avatar"`
	IsBlueVerified bool `json:"is_blue_verified"`
	Verification   struct {
		VerifiedType string `json:"verified_type"`
	} `json:"verification"`
	Privacy struct {
		Protected bool `json:"protected"`
	} `json:"privacy"`
	HasHiddenSubscriptions bool  `json:"has_hidden_subscriptions_on_profile"`
	PremiumGiftingEligible *bool `json:"premium_gifting_eligible"`
}

func parseProfile(body []byte) (domain.Profile, error) {
	var payload struct {
		Data struct {
			User *struct {
				Result *userResult `json:"result"`
			} `json:"user"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return domain.Profile{}, fmt.Errorf("%w: decode: %v", domain.ErrUnavailable, err)
	}
	if payload.Data.User == nil || payload.Data.User.Result == nil {
		if len(payload.Errors) > 0 {
			return domain.Profile{}, fmt.Errorf("%w: %s", domain.ErrUnavailable, payload.Errors[0].Message)
		}
		return domain.Profile{Found: false}, nil
	}
	u := payload.Data.User.Result
	if u.Typename == "UserUnavailable" {
		return domain.Profile{Found: true, Suspended: strings.EqualFold(u.Reason, "Suspended")}, nil
	}
	return domain.Profile{
		Found:               true,
		ScreenName:          u.Core.ScreenName,
		Name:                u.Core.Name,
		AvatarURL:           u.Avatar.ImageURL,
		IsBlueVerified:      u.IsBlueVerified,
		VerifiedType:        u.Verification.VerifiedType,
		Protected:           u.Privacy.Protected,
		HiddenSubscriptions: u.HasHiddenSubscriptions,
		GiftingEligible:     u.PremiumGiftingEligible,
	}, nil
}
