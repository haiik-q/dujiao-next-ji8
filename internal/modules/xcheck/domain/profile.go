package domain

import (
	"errors"
	"time"
)

// Profile 是从 X 读取到的、判断 Premium 赠礼资格所需的公开资料。
type Profile struct {
	Found               bool
	Suspended           bool
	ScreenName          string
	Name                string
	AvatarURL           string
	IsBlueVerified      bool
	VerifiedType        string // Business / Government 等机构认证；普通蓝标为空
	Protected           bool
	HiddenSubscriptions bool
	// GiftingEligible 是 X 直接返回的 premium_gifting_eligible；字段缺失时为 nil，由本地规则推断。
	GiftingEligible *bool
}

// 不符合赠礼资格的原因。
const (
	ReasonNotFound             = "not_found"
	ReasonSuspended            = "suspended"
	ReasonVerifiedOrganization = "verified_organization"
	ReasonAlreadyPremium       = "already_premium"
	ReasonProtected            = "protected"
	ReasonHiddenSubscriptions  = "hidden_subscriptions"
	ReasonNotEligible          = "not_eligible"
)

// ProfileView 是返回给前端的资料摘要。
type ProfileView struct {
	ScreenName     string `json:"screen_name"`
	Name           string `json:"name"`
	AvatarURL      string `json:"avatar_url,omitempty"`
	IsBlueVerified bool   `json:"is_blue_verified"`
	Protected      bool   `json:"protected"`
}

// Result 是一次 X ID 自检的结果。
type Result struct {
	Handle    string       `json:"handle"`
	Eligible  bool         `json:"eligible"`
	Reason    string       `json:"reason,omitempty"`
	Profile   *ProfileView `json:"profile,omitempty"`
	CheckedAt time.Time    `json:"checked_at"`
}

var (
	// ErrInvalidHandle 表示输入不是合法的 X 用户名。
	ErrInvalidHandle = errors.New("invalid x handle")
	// ErrUnavailable 表示没有可用的查询账号，或 X 暂时无法查询。
	ErrUnavailable = errors.New("x lookup unavailable")
)
