import type {
  AccountPayload,
  AccountEmailList,
  MailboxProviderPayload,
  MailboxProviderRecord,
  AccountRecord,
  RandomProfile,
  AuthUser,
  ApiEnvelope,
  ApiErrorEnvelope,
  CardDetail,
  CardEventView,
  CardImportPayload,
  CardRecord,
  ImportResult,
  LoginPayload,
  UserPayload,
  UserRecord,
} from "@/lib/types"

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api"

export class ApiRequestError extends Error {
  status: number
  code: string

  constructor(message: string, status: number, code: string) {
    super(message)
    this.name = "ApiRequestError"
    this.status = status
    this.code = code
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
  })

  if (response.status === 204) {
    return undefined as T
  }

  const payload = (await response.json()) as ApiEnvelope<T> | ApiErrorEnvelope
  if (!response.ok) {
    const errorCode = "error" in payload ? payload.error.code : "request_failed"
    const errorMessage = "error" in payload ? payload.error.message : "Request failed"
    throw new ApiRequestError(errorMessage, response.status, errorCode)
  }

  return (payload as ApiEnvelope<T>).data
}

export const api = {
  login(payload: LoginPayload) {
    return request<AuthUser>("/auth/login", {
      method: "POST",
      body: JSON.stringify(payload),
    })
  },
  getCurrentUser() {
    return request<AuthUser>("/auth/me")
  },
  logout() {
    return request<void>("/auth/logout", {
      method: "POST",
    })
  },
  getCards() {
    return request<CardRecord[]>("/cards")
  },
  importCards(payload: CardImportPayload) {
    return request<ImportResult>("/cards/import", {
      method: "POST",
      body: JSON.stringify(payload),
    })
  },
  getCard(cardId: string) {
    return request<CardDetail>(`/cards/${cardId}`)
  },
  activateCard(cardId: number) {
    return request<CardEventView>(`/cards/${cardId}/activate`, {
      method: "POST",
    })
  },
  queryCard(cardId: number) {
    return request<CardEventView>(`/cards/${cardId}/query`, {
      method: "POST",
    })
  },
  getBilling(cardId: number) {
    return request<CardEventView>(`/cards/${cardId}/billing`, {
      method: "POST",
    })
  },
  getThreeDS(cardId: number, minutes: number) {
    return request<CardEventView>(`/cards/${cardId}/3ds`, {
      method: "POST",
      body: JSON.stringify({ minutes }),
    })
  },
  refreshProfile(cardId: number) {
    return request<CardEventView>(`/cards/${cardId}/profile/refresh`, {
      method: "POST",
    })
  },
  deleteCard(cardId: number) {
    return request<void>(`/cards/${cardId}`, {
      method: "DELETE",
    })
  },
  getAccounts() {
    return request<AccountRecord[]>("/accounts")
  },
  getAccountEmails(accountId: number) {
    return request<AccountEmailList>(`/accounts/${accountId}/emails`)
  },
  getRandomProfile() {
    return request<RandomProfile>("/profiles/random")
  },
  getMailboxProviders() {
    return request<MailboxProviderRecord[]>("/mailbox-providers")
  },
  createMailboxProvider(payload: MailboxProviderPayload) {
    return request<MailboxProviderRecord>("/mailbox-providers", {
      method: "POST",
      body: JSON.stringify(payload),
    })
  },
  updateMailboxProvider(providerId: number, payload: MailboxProviderPayload) {
    return request<MailboxProviderRecord>(`/mailbox-providers/${providerId}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    })
  },
  deleteMailboxProvider(providerId: number) {
    return request<void>(`/mailbox-providers/${providerId}`, {
      method: "DELETE",
    })
  },
  createAccount(payload: AccountPayload) {
    return request<AccountRecord>("/accounts", {
      method: "POST",
      body: JSON.stringify(payload),
    })
  },
  updateAccount(accountId: number, payload: AccountPayload) {
    return request<AccountRecord>(`/accounts/${accountId}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    })
  },
  deleteAccount(accountId: number) {
    return request<void>(`/accounts/${accountId}`, {
      method: "DELETE",
    })
  },
  getWarranties(accountId: number) {
    return request<AccountRecord[]>(`/accounts/${accountId}/warranties`)
  },
  createWarranty(accountId: number, payload: AccountPayload) {
    return request<AccountRecord>(`/accounts/${accountId}/warranties`, {
      method: "POST",
      body: JSON.stringify(payload),
    })
  },
  updateWarranty(accountId: number, warrantyId: number, payload: AccountPayload) {
    return request<AccountRecord>(
      `/accounts/${accountId}/warranties/${warrantyId}`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      }
    )
  },
  deleteWarranty(accountId: number, warrantyId: number) {
    return request<void>(`/accounts/${accountId}/warranties/${warrantyId}`, {
      method: "DELETE",
    })
  },
  getUsers() {
    return request<UserRecord[]>("/users")
  },
  createUser(payload: UserPayload) {
    return request<UserRecord>("/users", {
      method: "POST",
      body: JSON.stringify(payload),
    })
  },
  updateUser(userId: number, payload: UserPayload) {
    return request<UserRecord>(`/users/${userId}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    })
  },
  deleteUser(userId: number) {
    return request<void>(`/users/${userId}`, {
      method: "DELETE",
    })
  },
}
