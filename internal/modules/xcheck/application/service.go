// Package application 实现 X Premium 赠礼资格自检：规范化用户名、查询资料、判定原因并缓存结果。
package application

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dujiao-next/internal/modules/xcheck/domain"
)

// CacheTTL 同一用户名的结果缓存时长：客人反复点击不消耗查询账号额度。
const CacheTTL = 10 * time.Minute

var handlePattern = regexp.MustCompile(`^[A-Za-z0-9_]{1,15}$`)

// ProfileLookup 查询 X 用户公开资料。
type ProfileLookup interface {
	Lookup(ctx context.Context, handle string) (domain.Profile, error)
}

type cacheEntry struct {
	result  domain.Result
	expires time.Time
}

// Service X ID 自检服务，并发安全。
type Service struct {
	lookup ProfileLookup
	now    func() time.Time

	mu    sync.Mutex
	cache map[string]cacheEntry
}

// NewService 创建自检服务。
func NewService(lookup ProfileLookup) *Service {
	return &Service{lookup: lookup, now: time.Now, cache: make(map[string]cacheEntry)}
}

// NormalizeHandle 接受 "@name"、"name" 或 x.com / twitter.com 的主页链接，返回用户名。
func NormalizeHandle(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	lower := strings.ToLower(s)
	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(lower, prefix) {
			s, lower = s[len(prefix):], lower[len(prefix):]
		}
	}
	for _, host := range []string{"www.x.com/", "x.com/", "mobile.x.com/", "www.twitter.com/", "twitter.com/", "mobile.twitter.com/"} {
		if strings.HasPrefix(lower, host) {
			s = s[len(host):]
			break
		}
	}
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimPrefix(s, "@")
	if !handlePattern.MatchString(s) {
		return "", domain.ErrInvalidHandle
	}
	return s, nil
}

// Check 自检一个 X 用户名是否可以接收 Premium 赠礼。
func (s *Service) Check(ctx context.Context, raw string) (domain.Result, error) {
	handle, err := NormalizeHandle(raw)
	if err != nil {
		return domain.Result{}, err
	}
	key := strings.ToLower(handle)
	now := s.now()

	s.mu.Lock()
	if entry, ok := s.cache[key]; ok && now.Before(entry.expires) {
		s.mu.Unlock()
		return entry.result, nil
	}
	s.mu.Unlock()

	profile, err := s.lookup.Lookup(ctx, handle)
	if err != nil {
		return domain.Result{}, err
	}
	result := Evaluate(handle, profile, now)
	s.store(key, result, now)
	return result, nil
}

// Verify 下单时的强制核实：不读缓存、直接问 X（结果写回缓存）。不符合返回 ErrNotEligible；
// X 查不了时，若 CacheTTL 内查到过「可以接收」仍放行（兑换站兑换和付款前还会再查），否则返回 ErrUnavailable。
func (s *Service) Verify(ctx context.Context, raw string) error {
	handle, err := NormalizeHandle(raw)
	if err != nil {
		return err
	}
	key := strings.ToLower(handle)
	profile, lookupErr := s.lookup.Lookup(ctx, handle)
	now := s.now()
	if lookupErr != nil {
		s.mu.Lock()
		entry, ok := s.cache[key]
		s.mu.Unlock()
		if ok && now.Before(entry.expires) && entry.result.Eligible {
			return nil
		}
		return fmt.Errorf("%w: %v", domain.ErrUnavailable, lookupErr)
	}
	result := Evaluate(handle, profile, now)
	s.store(key, result, now)
	if !result.Eligible {
		return fmt.Errorf("%w: @%s %s", domain.ErrNotEligible, result.Handle, result.Reason)
	}
	return nil
}

func (s *Service) store(key string, result domain.Result, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, e := range s.cache {
		if !now.Before(e.expires) {
			delete(s.cache, k)
		}
	}
	s.cache[key] = cacheEntry{result: result, expires: now.Add(CacheTTL)}
}

// Evaluate 根据资料判定资格。以 X 返回的 premium_gifting_eligible 为准，原因按常见程度依次判断。
func Evaluate(handle string, p domain.Profile, now time.Time) domain.Result {
	result := domain.Result{Handle: handle, CheckedAt: now.UTC()}
	switch {
	case !p.Found:
		result.Reason = domain.ReasonNotFound
		return result
	case p.Suspended:
		result.Reason = domain.ReasonSuspended
		return result
	}
	if p.ScreenName != "" {
		result.Handle = p.ScreenName
	}
	result.Profile = &domain.ProfileView{
		ScreenName:     result.Handle,
		Name:           p.Name,
		AvatarURL:      p.AvatarURL,
		IsBlueVerified: p.IsBlueVerified,
		Protected:      p.Protected,
	}
	reason := ""
	switch {
	case p.VerifiedType != "":
		reason = domain.ReasonVerifiedOrganization
	case p.IsBlueVerified:
		reason = domain.ReasonAlreadyPremium
	case p.Protected:
		reason = domain.ReasonProtected
	case p.HiddenSubscriptions:
		reason = domain.ReasonHiddenSubscriptions
	}
	eligible := reason == ""
	if p.GiftingEligible != nil {
		eligible = *p.GiftingEligible
		if !eligible && reason == "" {
			reason = domain.ReasonNotEligible
		}
	}
	result.Eligible = eligible
	if !eligible {
		result.Reason = reason
	}
	return result
}
