import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { cleanup, render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import { MessageProvider } from "@/components/message"
import { SubAccountDialog } from "@/features/accounts/sub-account-dialog"
import { api } from "@/lib/api"
import type { AccountRecord } from "@/lib/types"

vi.mock("@/lib/api", () => ({
  api: {
    getSubAccounts: vi.fn(),
    createSubAccount: vi.fn(),
    updateSubAccount: vi.fn(),
    deleteSubAccount: vi.fn(),
    getAccountEmails: vi.fn(),
    getMailboxProviders: vi.fn(),
  },
}))

vi.mock("@/features/auth/auth-provider", () => ({
  useAuth: () => ({
    user: { id: 1, username: "admin", role: "admin" },
  }),
}))

const mockedApi = vi.mocked(api)

const parentAccount: AccountRecord = {
  id: 12,
  account: "owner@codex.test",
  password: "secret-pass",
  maskedPassword: "s********s",
  type: "codex",
  startTime: "2026-03-20T08:00:00.000Z",
  endTime: "2026-04-20T08:00:00.000Z",
  status: "normal",
  remark: "codex parent",
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
}

function renderDialog() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })

  render(
    <QueryClientProvider client={queryClient}>
      <MessageProvider>
        <SubAccountDialog open onOpenChange={vi.fn()} account={parentAccount} />
      </MessageProvider>
    </QueryClientProvider>
  )
}

describe("SubAccountDialog", () => {
  beforeEach(() => {
    mockedApi.getSubAccounts.mockReset()
    mockedApi.createSubAccount.mockReset()
    mockedApi.updateSubAccount.mockReset()
    mockedApi.deleteSubAccount.mockReset()
    mockedApi.getAccountEmails.mockReset()
    mockedApi.getMailboxProviders.mockReset()
    mockedApi.getSubAccounts.mockResolvedValue([])
    mockedApi.getMailboxProviders.mockResolvedValue([])
  })

  afterEach(() => {
    cleanup()
  })

  it("quick creates a sub account with the parent suffix and password", async () => {
    const user = userEvent.setup()
    mockedApi.createSubAccount.mockResolvedValue({
      ...parentAccount,
      id: 99,
      account: "owner-aaaa@codex.test",
    })

    renderDialog()

    await waitFor(() => {
      expect(mockedApi.getSubAccounts).toHaveBeenCalledWith(parentAccount.id)
    })

    await user.click(screen.getByRole("button", { name: "一键生成子号" }))

    await waitFor(() => {
      expect(mockedApi.createSubAccount).toHaveBeenCalledWith(
        parentAccount.id,
        expect.objectContaining({
          password: parentAccount.password,
          type: "codex",
          status: "normal",
          remark: "",
          createMailbox: true,
          useServerTimeSchedule: true,
        })
      )
    })

    const payload = mockedApi.createSubAccount.mock.calls[0]?.[1]
    expect(payload?.account).toMatch(/^[a-z]{8,10}@codex\.test$/)
    expect(payload?.startTime).toBeUndefined()
    expect(payload?.endTime).toBeUndefined()
  })

  it("prefills manual creation with the parent password and suffix", async () => {
    const user = userEvent.setup()

    renderDialog()

    await waitFor(() => {
      expect(mockedApi.getSubAccounts).toHaveBeenCalledWith(parentAccount.id)
    })

    await user.click(screen.getByRole("button", { name: "新增子号" }))

    const accountInput = screen.getByLabelText("账号") as HTMLInputElement
    const passwordInput = screen.getByLabelText("密码") as HTMLInputElement

    expect(accountInput.value).toMatch(/^[a-z]{8,10}@codex\.test$/)
    expect(passwordInput.value).toBe(parentAccount.password)
    expect(screen.getByLabelText("类型")).toHaveValue("codex")
  })
})
