import { cleanup, render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { afterEach, describe, expect, it, vi } from "vitest"

import { AccountTable } from "@/features/accounts/account-table"
import type { AccountRecord } from "@/lib/types"

const accounts: AccountRecord[] = [
  {
    id: 1,
    account: "blocked@example.com",
    password: "secret-1",
    maskedPassword: "s******1",
    type: "plus",
    status: "blocked",
    remark: "need warranty",
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 2,
    account: "normal@example.com",
    password: "secret-2",
    maskedPassword: "s******2",
    type: "business",
    status: "normal",
    remark: "",
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 3,
    account: "codex@example.com",
    password: "secret-3",
    maskedPassword: "s******3",
    type: "codex",
    status: "normal",
    remark: "sub accounts",
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
]

afterEach(() => {
  cleanup()
})

describe("AccountTable", () => {
  it("centers the empty state", () => {
    render(<AccountTable accounts={[]} />)

    expect(screen.getByText("当前列表为空。")).toHaveClass(
      "flex",
      "items-center",
      "justify-center",
      "text-center"
    )
  })

  it("shows the warranty action only for blocked accounts", () => {
    render(
      <AccountTable
        accounts={accounts}
        onEdit={vi.fn()}
        onDelete={vi.fn().mockResolvedValue(undefined)}
        onOpenWarranty={vi.fn()}
      />
    )

    expect(screen.getAllByRole("button", { name: /质保/ })).toHaveLength(1)
  })

  it("shows the sub account action only for codex accounts", () => {
    render(
      <AccountTable
        accounts={accounts}
        onEdit={vi.fn()}
        onDelete={vi.fn().mockResolvedValue(undefined)}
        onOpenSubAccounts={vi.fn()}
      />
    )

    expect(screen.getAllByRole("button", { name: /子号管理/ })).toHaveLength(1)
  })

  it("reveals the password when toggled", async () => {
    const user = userEvent.setup()

    render(
      <AccountTable
        accounts={[accounts[0]]}
        onEdit={vi.fn()}
        onDelete={vi.fn().mockResolvedValue(undefined)}
        onOpenWarranty={vi.fn()}
      />
    )

    expect(screen.getAllByText("s******1")[0]).toBeInTheDocument()
    await user.click(screen.getByRole("button", { name: "显示密码" }))
    expect(screen.getByText("secret-1")).toBeInTheDocument()
  })

  it("renders account and password as copyable buttons", () => {
    render(
      <AccountTable
        accounts={[accounts[0]]}
        onEdit={vi.fn()}
        onDelete={vi.fn().mockResolvedValue(undefined)}
        onOpenWarranty={vi.fn()}
      />
    )

    expect(screen.getByRole("button", { name: "blocked@example.com" })).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "s******1" })).toBeInTheDocument()
  })

  it("renders the email action when email callback is provided", async () => {
    const user = userEvent.setup()
    const onOpenEmails = vi.fn()

    render(
      <AccountTable
        accounts={[accounts[0]]}
        onEdit={vi.fn()}
        onDelete={vi.fn().mockResolvedValue(undefined)}
        onOpenEmails={onOpenEmails}
        onOpenWarranty={vi.fn()}
      />
    )

    await user.click(screen.getByRole("button", { name: /邮件/ }))
    expect(onOpenEmails).toHaveBeenCalledWith(accounts[0])
  })
})
