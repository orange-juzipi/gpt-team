import { cleanup, render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { afterEach, describe, expect, it, vi } from "vitest"

import { MailboxProviderFormDialog } from "@/features/mailboxes/mailbox-provider-form-dialog"

afterEach(() => {
  cleanup()
})

describe("MailboxProviderFormDialog", () => {
  it("toggles mailbox password visibility", async () => {
    const user = userEvent.setup()

    render(
      <MailboxProviderFormDialog
        open
        onOpenChange={vi.fn()}
        title="编辑邮箱"
        description="test"
        submitLabel="保存修改"
        isPending={false}
        onSubmit={vi.fn()}
        initialValue={{
          id: 1,
          providerType: "cloudmail",
          domainSuffix: "mail.example",
          accountEmail: "admin@mail.example",
          password: "secret-password",
          maskedPassword: "s************d",
          remark: "",
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        }}
      />
    )

    const passwordInput = screen.getByLabelText("邮箱密码") as HTMLInputElement
    expect(passwordInput).toHaveAttribute("type", "password")

    await user.click(screen.getByRole("button", { name: "显示邮箱密码" }))
    expect(passwordInput).toHaveAttribute("type", "text")

    await user.click(screen.getByRole("button", { name: "隐藏邮箱密码" }))
    expect(passwordInput).toHaveAttribute("type", "password")
  })
})
