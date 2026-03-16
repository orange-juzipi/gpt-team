import type {
  CardLimit,
  AccountStatus,
  AccountType,
  CardEventType,
  CardStatus,
  CardType,
  MailboxProviderType,
  UserRole,
} from "@/lib/types"

const dateTimeFormatter = new Intl.DateTimeFormat("zh-CN", {
  dateStyle: "medium",
  timeStyle: "short",
})

export function formatDateTime(value?: string) {
  if (!value) {
    return "未记录"
  }

  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    return value
  }

  return dateTimeFormatter.format(parsed)
}

export function toDateTimeLocalValue(value?: string) {
  if (!value) {
    return ""
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ""
  }

  const offset = date.getTimezoneOffset()
  const localDate = new Date(date.getTime() - offset * 60_000)
  return localDate.toISOString().slice(0, 16)
}

export function fromDateTimeLocalValue(value: string) {
  if (!value) {
    return undefined
  }

  return new Date(value).toISOString()
}

export function cardStatusLabel(status: CardStatus) {
  return status === "activated" ? "已激活" : "未激活"
}

export function cardTypeLabel(type: CardType) {
  switch (type) {
    case "uk":
      return "英卡"
    case "es":
      return "西班牙卡"
    case "us":
    default:
      return "美卡"
  }
}

export function cardCountryLabel(type: CardType) {
  switch (type) {
    case "uk":
      return "英国"
    case "es":
      return "西班牙"
    case "us":
    default:
      return "美国"
  }
}

export function cardLimitLabel(limit: CardLimit | number) {
  return `${limit} 刀`
}

export function accountStatusLabel(status: AccountStatus) {
  return status === "blocked" ? "已封" : "正常"
}

export function accountTypeLabel(type: AccountType) {
  return type === "business" ? "Business" : "Plus"
}

export function eventTypeLabel(type: CardEventType) {
  switch (type) {
    case "activate":
      return "激活"
    case "query":
      return "查询"
    case "billing":
      return "账单"
    case "three_ds":
      return "3DS"
    case "identity_refresh":
      return "身份资料"
    default:
      return type
  }
}

export function userRoleLabel(role: UserRole) {
  return role === "admin" ? "管理员" : "普通用户"
}

export function mailboxProviderTypeLabel(type: MailboxProviderType) {
  return type === "duckmail" ? "DuckMail" : "Cloudmail"
}
