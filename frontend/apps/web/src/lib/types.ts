export type CardStatus = "unactivated" | "activated"
export type CardType = "us" | "uk" | "es"
export type CardLimit = 0 | 1 | 2
export type CardEventType =
  | "activate"
  | "query"
  | "billing"
  | "three_ds"
  | "identity_refresh"

export type AccountType = "plus" | "business" | "codex"
export type AccountStatus = "normal" | "blocked"
export type UserRole = "admin" | "member"
export type MailboxProviderType = "cloudmail" | "duckmail"

export type CardRecord = {
  id: number
  code: string
  cardType: CardType
  cardLimit: CardLimit
  status: CardStatus
  remoteStatus: string
  remoteCardId?: number
  lastFour: string
  expiryDate: string
  fullName: string
  birthday: string
  streetAddress: string
  district: string
  city: string
  state: string
  stateFull: string
  zipCode: string
  phoneNumber: string
  lastSyncedAt?: string
  createdAt: string
  updatedAt: string
}

export type CardEventView = {
  id: number
  type: CardEventType
  success: boolean
  errorMessage?: string
  createdAt: string
  data?: unknown
}

export type CardDetail = {
  card: CardRecord
  latestActivation?: CardEventView
  latestQuery?: CardEventView
  latestBilling?: CardEventView
  latestThreeDS?: CardEventView
  latestIdentity?: CardEventView
}

export type ImportResult = {
  createdCount: number
  duplicates: string[]
  items: CardRecord[]
}

export type CardImportPayload = {
  rawText: string
  cardType: CardType
  cardLimit: CardLimit
}

export type AccountRecord = {
  id: number
  account: string
  password: string
  maskedPassword: string
  type: AccountType
  startTime?: string
  endTime?: string
  status: AccountStatus
  remark: string
  parentId?: number
  createdAt: string
  updatedAt: string
}

export type AccountPayload = {
  account: string
  password: string
  type: AccountType
  startTime?: string
  endTime?: string
  status: AccountStatus
  remark: string
  createMailbox?: boolean
  useServerTimeSchedule?: boolean
}

export type AccountEmailRecord = {
  id: string
  account: string
  from: string
  fromName: string
  subject: string
  preview: string
  receivedAt: string
}

export type AccountEmailList = {
  accountId: number
  account: string
  items: AccountEmailRecord[]
}

export type RandomProfile = {
  fullName: string
  birthday: string
}

export type MailboxProviderRecord = {
  id: number
  providerType: MailboxProviderType
  domainSuffix: string
  accountEmail: string
  password: string
  maskedPassword: string
  remark: string
  createdAt: string
  updatedAt: string
}

export type MailboxProviderPayload = {
  providerType: MailboxProviderType
  domainSuffix: string
  accountEmail: string
  password: string
  remark: string
}

export type UserRecord = {
  id: number
  username: string
  role: UserRole
  createdAt: string
  updatedAt: string
}

export type AuthUser = UserRecord

export type LoginPayload = {
  username: string
  password: string
}

export type UserPayload = {
  username: string
  password: string
  role: UserRole
}

export type ApiEnvelope<T> = {
  data: T
}

export type ApiErrorEnvelope = {
  error: {
    code: string
    message: string
  }
}
