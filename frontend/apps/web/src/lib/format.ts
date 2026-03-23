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

const BEIJING_TIME_ZONE = "Asia/Shanghai"
const BEIJING_OFFSET_HOURS = 8

const dateTimeFormatter = new Intl.DateTimeFormat("zh-CN", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: BEIJING_TIME_ZONE,
})

type DateTimeInputTimeZone = "local" | "utc" | "beijing"

export function formatDateTime(
  value?: string,
  options?: {
    inputTimeZone?: DateTimeInputTimeZone
  }
) {
  if (!value) {
    return "未记录"
  }

  const parsed = parseDateTime(value, options?.inputTimeZone ?? "local")
  if (Number.isNaN(parsed.getTime())) {
    return value
  }

  return dateTimeFormatter.format(parsed)
}

export function toDateTimeLocalValue(value?: string) {
  if (!value) {
    return ""
  }

  const date = parseDateTime(value, "local")
  if (Number.isNaN(date.getTime())) {
    return ""
  }

  const beijingDate = new Date(date.getTime() + BEIJING_OFFSET_HOURS * 60 * 60_000)

  return [
    beijingDate.getUTCFullYear(),
    pad(beijingDate.getUTCMonth() + 1),
    pad(beijingDate.getUTCDate()),
  ].join("-") + `T${pad(beijingDate.getUTCHours())}:${pad(beijingDate.getUTCMinutes())}`
}

export function fromDateTimeLocalValue(value: string) {
  if (!value) {
    return undefined
  }

  const match = value.match(
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2}))?$/
  )
  if (!match) {
    return undefined
  }

  const [, year, month, day, hour, minute, second = "00"] = match
  const timestamp = Date.UTC(
    Number(year),
    Number(month) - 1,
    Number(day),
    Number(hour) - BEIJING_OFFSET_HOURS,
    Number(minute),
    Number(second)
  )

  return new Date(timestamp).toISOString()
}

function parseDateTime(value: string, inputTimeZone: DateTimeInputTimeZone) {
  const trimmed = value.trim()
  if (trimmed === "") {
    return new Date(Number.NaN)
  }

  if (hasExplicitTimeZone(trimmed)) {
    return new Date(trimmed)
  }

  const normalized = normalizeDateTime(trimmed)

  switch (inputTimeZone) {
    case "utc":
      return new Date(`${normalized}Z`)
    case "beijing":
      return new Date(`${normalized}+08:00`)
    case "local":
    default:
      return new Date(trimmed)
  }
}

function hasExplicitTimeZone(value: string) {
  return /(?:[zZ]|[+-]\d{2}:\d{2}|[+-]\d{4})$/.test(value)
}

function normalizeDateTime(value: string) {
  return value.includes("T") ? value : value.replace(" ", "T")
}

function pad(value: number) {
  return String(value).padStart(2, "0")
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
  switch (type) {
    case "business":
      return "Business"
    case "codex":
      return "Codex"
    case "plus":
    default:
      return "Plus"
  }
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
