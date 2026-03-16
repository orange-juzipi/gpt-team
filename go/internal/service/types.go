package service

import (
	"time"

	"gpt-team-api/internal/integration/mailbox"
	"gpt-team-api/internal/model"
)

type CardRecord struct {
	ID            uint64           `json:"id"`
	Code          string           `json:"code"`
	CardType      model.CardType   `json:"cardType"`
	CardLimit     int              `json:"cardLimit"`
	Status        model.CardStatus `json:"status"`
	RemoteStatus  string           `json:"remoteStatus"`
	RemoteCardID  *uint64          `json:"remoteCardId,omitempty"`
	LastFour      string           `json:"lastFour"`
	ExpiryDate    string           `json:"expiryDate"`
	FullName      string           `json:"fullName"`
	Birthday      string           `json:"birthday"`
	StreetAddress string           `json:"streetAddress"`
	District      string           `json:"district"`
	City          string           `json:"city"`
	State         string           `json:"state"`
	StateFull     string           `json:"stateFull"`
	ZipCode       string           `json:"zipCode"`
	PhoneNumber   string           `json:"phoneNumber"`
	LastSyncedAt  *time.Time       `json:"lastSyncedAt,omitempty"`
	CreatedAt     time.Time        `json:"createdAt"`
	UpdatedAt     time.Time        `json:"updatedAt"`
}

type CardEventView struct {
	ID           uint64              `json:"id"`
	Type         model.CardEventType `json:"type"`
	Success      bool                `json:"success"`
	ErrorMessage string              `json:"errorMessage,omitempty"`
	CreatedAt    time.Time           `json:"createdAt"`
	Data         any                 `json:"data,omitempty"`
}

type CardDetail struct {
	Card             CardRecord     `json:"card"`
	LatestActivation *CardEventView `json:"latestActivation,omitempty"`
	LatestQuery      *CardEventView `json:"latestQuery,omitempty"`
	LatestBilling    *CardEventView `json:"latestBilling,omitempty"`
	LatestThreeDS    *CardEventView `json:"latestThreeDS,omitempty"`
	LatestIdentity   *CardEventView `json:"latestIdentity,omitempty"`
}

type ImportResult struct {
	CreatedCount int          `json:"createdCount"`
	Duplicates   []string     `json:"duplicates"`
	Items        []CardRecord `json:"items"`
}

type ImportCardsInput struct {
	RawText   string
	CardType  model.CardType
	CardLimit int
}

type AccountRecord struct {
	ID             uint64              `json:"id"`
	Account        string              `json:"account"`
	Password       string              `json:"password"`
	MaskedPassword string              `json:"maskedPassword"`
	Type           model.AccountType   `json:"type"`
	StartTime      *time.Time          `json:"startTime,omitempty"`
	EndTime        *time.Time          `json:"endTime,omitempty"`
	Status         model.AccountStatus `json:"status"`
	Remark         string              `json:"remark"`
	ParentID       *uint64             `json:"parentId,omitempty"`
	CreatedAt      time.Time           `json:"createdAt"`
	UpdatedAt      time.Time           `json:"updatedAt"`
}

type AccountInput struct {
	Account       string
	Password      string
	Type          model.AccountType
	StartTime     *time.Time
	EndTime       *time.Time
	Status        model.AccountStatus
	Remark        string
	CreateMailbox bool
}

type AccountEmailRecord struct {
	ID         string `json:"id"`
	Account    string `json:"account"`
	From       string `json:"from"`
	FromName   string `json:"fromName"`
	Subject    string `json:"subject"`
	Preview    string `json:"preview"`
	ReceivedAt string `json:"receivedAt"`
}

type AccountEmailList struct {
	AccountID uint64               `json:"accountId"`
	Account   string               `json:"account"`
	Items     []AccountEmailRecord `json:"items"`
}

type MailboxProviderRecord struct {
	ID             uint64                    `json:"id"`
	ProviderType   model.MailboxProviderType `json:"providerType"`
	DomainSuffix   string                    `json:"domainSuffix"`
	AccountEmail   string                    `json:"accountEmail"`
	Password       string                    `json:"password"`
	MaskedPassword string                    `json:"maskedPassword"`
	Remark         string                    `json:"remark"`
	CreatedAt      time.Time                 `json:"createdAt"`
	UpdatedAt      time.Time                 `json:"updatedAt"`
}

type MailboxProviderInput struct {
	ProviderType model.MailboxProviderType
	DomainSuffix string
	AccountEmail string
	Password     string
	Remark       string
}

type UserRecord struct {
	ID        uint64         `json:"id"`
	Username  string         `json:"username"`
	Role      model.UserRole `json:"role"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type UserInput struct {
	Username string
	Password string
	Role     model.UserRole
}

type LoginInput struct {
	Username string
	Password string
}

type RandomProfile struct {
	FullName string `json:"fullName"`
	Birthday string `json:"birthday"`
}

type AuthLoginResult struct {
	User  UserRecord `json:"user"`
	Token string     `json:"-"`
}

func toAccountEmailRecord(item mailbox.Email) AccountEmailRecord {
	return AccountEmailRecord{
		ID:         item.ID,
		Account:    item.Account,
		From:       item.From,
		FromName:   item.FromName,
		Subject:    item.Subject,
		Preview:    item.Preview,
		ReceivedAt: item.ReceivedAt,
	}
}
