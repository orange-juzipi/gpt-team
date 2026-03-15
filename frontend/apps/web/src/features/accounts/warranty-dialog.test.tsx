import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { cleanup, render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import { MessageProvider } from "@/components/message"
import { WarrantyDialog } from "@/features/accounts/warranty-dialog"
import { api } from "@/lib/api"
import { fromDateTimeLocalValue } from "@/lib/format"
import type { AccountRecord } from "@/lib/types"

vi.mock("@/lib/api", () => ({
  api: {
    getWarranties: vi.fn(),
    getAccountEmails: vi.fn(),
    getMailboxProviders: vi.fn(),
    createWarranty: vi.fn(),
    updateWarranty: vi.fn(),
    deleteWarranty: vi.fn(),
  },
}))

vi.mock("@/features/auth/auth-provider", () => ({
  useAuth: () => ({
    user: { id: 1, username: "admin", role: "admin" },
  }),
}))

const mockedApi = vi.mocked(api)

const parentAccount: AccountRecord = {
  id: 7,
  account: "blocked@example.com",
  password: "secret",
  maskedPassword: "s****t",
  type: "plus",
  status: "blocked",
  remark: "blocked",
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
}

describe("WarrantyDialog", () => {
  beforeEach(() => {
    mockedApi.getWarranties.mockReset()
    mockedApi.getAccountEmails.mockReset()
    mockedApi.getMailboxProviders.mockReset()
    mockedApi.createWarranty.mockReset()
    mockedApi.updateWarranty.mockReset()
    mockedApi.deleteWarranty.mockReset()
    mockedApi.getMailboxProviders.mockResolvedValue([])
  })

  afterEach(() => {
    cleanup()
  })

  it("creates a warranty account and refreshes the list query", async () => {
    const user = userEvent.setup()
    mockedApi.getWarranties.mockResolvedValue([])
    mockedApi.createWarranty.mockResolvedValue({
      ...parentAccount,
      id: 9,
      account: "warranty@example.com",
      password: "newpass",
      maskedPassword: "n*****s",
      status: "normal",
    })

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
          <WarrantyDialog open onOpenChange={vi.fn()} account={parentAccount} />
        </MessageProvider>
      </QueryClientProvider>
    )

    await waitFor(() => {
      expect(mockedApi.getWarranties).toHaveBeenCalledWith(parentAccount.id)
    })

    await user.click(screen.getByRole("button", { name: "新增质保账号" }))
    expect(screen.getByRole("button", { name: "随机账号" })).toBeDisabled()
    await user.type(screen.getByLabelText("账号"), "warranty@example.com")
    await user.type(screen.getByLabelText("密码"), "newpass")
    await user.type(screen.getByLabelText("备注"), "backup account")

    const startTimeInput = screen.getByLabelText("开始时间") as HTMLInputElement
    const endTimeInput = screen.getByLabelText("结束时间") as HTMLInputElement

    expect(screen.getByLabelText("类型")).toHaveValue("business")
    expect(startTimeInput.value).not.toBe("")
    expect(endTimeInput.value).not.toBe("")

    const expectedStartTime = fromDateTimeLocalValue(startTimeInput.value)
    const expectedEndTime = fromDateTimeLocalValue(endTimeInput.value)
    const computedEndTime = new Date(expectedStartTime!)
    computedEndTime.setMonth(computedEndTime.getMonth() + 1)

    expect(expectedEndTime).toBe(computedEndTime.toISOString())

    await user.click(screen.getByRole("button", { name: "创建质保账号" }))

    await waitFor(() => {
      expect(mockedApi.createWarranty).toHaveBeenCalledWith(parentAccount.id, {
        account: "warranty@example.com",
        password: "newpass",
        type: "business",
        startTime: expectedStartTime,
        endTime: expectedEndTime,
        status: "normal",
        remark: "backup account",
        createMailbox: true,
      })
    })

    await waitFor(() => {
      expect(mockedApi.getWarranties.mock.calls.length).toBeGreaterThan(1)
    })
  })

  it("generates a random account with the selected mailbox suffix", async () => {
    const user = userEvent.setup()
    mockedApi.getWarranties.mockResolvedValue([])
    mockedApi.getMailboxProviders.mockResolvedValue([
      {
        id: 1,
        providerType: "cloudmail",
        domainSuffix: "alpha.test",
        accountEmail: "admin@alpha.test",
        password: "secret",
        maskedPassword: "s****t",
        remark: "",
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      {
        id: 2,
        providerType: "duckmail",
        domainSuffix: "beta.test",
        accountEmail: "",
        password: "dk_secret",
        maskedPassword: "d*****t",
        remark: "",
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ])

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
          <WarrantyDialog open onOpenChange={vi.fn()} account={parentAccount} />
        </MessageProvider>
      </QueryClientProvider>
    )

    await waitFor(() => {
      expect(mockedApi.getWarranties).toHaveBeenCalledWith(parentAccount.id)
    })

    await user.click(screen.getByRole("button", { name: "新增质保账号" }))
    await user.selectOptions(screen.getByLabelText("随机邮箱后缀"), "beta.test")
    await user.click(screen.getByRole("button", { name: "随机账号" }))

    const accountInput = screen.getByLabelText("账号") as HTMLInputElement
    expect(accountInput.value).toMatch(/^[a-z]{5,8}@beta\.test$/)
  })

  it("opens emails for a warranty account", async () => {
    const user = userEvent.setup()
    mockedApi.getWarranties.mockResolvedValue([
      {
        ...parentAccount,
        id: 11,
        account: "warranty@example.com",
        password: "secret2",
        maskedPassword: "s*****2",
        status: "normal",
      },
    ])
    mockedApi.getAccountEmails.mockResolvedValue({
      accountId: 11,
      account: "warranty@example.com",
      items: [],
    })

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
          <WarrantyDialog open onOpenChange={vi.fn()} account={parentAccount} />
        </MessageProvider>
      </QueryClientProvider>
    )

    await waitFor(() => {
      expect(mockedApi.getWarranties).toHaveBeenCalledWith(parentAccount.id)
    })

    await user.click(await screen.findByRole("button", { name: /邮件/i }))

    await waitFor(() => {
      expect(mockedApi.getAccountEmails).toHaveBeenCalledWith(11)
    })

    expect(screen.getByText("邮件记录")).toBeInTheDocument()
    expect(screen.getByText("当前邮箱：warranty@example.com")).toBeInTheDocument()
  })
})
