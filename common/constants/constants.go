package constants

const RoleAdmin = "admin"
const RoleDeveloper = "developer"
const RoleUser = "user"

// GPT pricing charge rate per token
const (
	GPT3CompletionCharge = 0.002 / 1000
	GPT3PromptCharge     = 0.002 / 1000
)

const (
	GPT4CompletionCharge = 0.06 / 1000
	GPT4PromptCharge     = 0.03 / 1000
)

const DollarToChineseCentsRate = 1100

const (
	RechargingCardActive   = "active"
	RechargingCardInactive = "inactive"
	RechargingCardUsed     = "used"
)

const (
	TransactionTypeRecharge = "recharge"
	TransactionTypeAdmin    = "admin"
)
const (
	Version = "1.0.0"
)

const (
	MaxRooms = 1000
)

// redis的key前缀
const (
	RedisKeyDocuments = "documents:"
	RedisKeyAiPrompt  = "ai-prompt:"
)

// meilisearch的集合key管理
const (
	MeiliIndexDocuments = "documents"
)
