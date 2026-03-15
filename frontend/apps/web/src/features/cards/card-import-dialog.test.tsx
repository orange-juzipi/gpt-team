import { cleanup, render, screen, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { afterEach, describe, expect, it, vi } from "vitest"

import { CardImportDialog } from "@/features/cards/card-import-dialog"

afterEach(() => {
  cleanup()
})

describe("CardImportDialog", () => {
  it("defaults to 英卡 and 0 刀", () => {
    render(
      <CardImportDialog
        open
        onOpenChange={vi.fn()}
        onSubmit={vi.fn().mockResolvedValue(undefined)}
        isPending={false}
      />
    )

    expect(screen.getByLabelText("卡片类型")).toHaveValue("uk")
    expect(screen.getByLabelText("额度")).toHaveValue("0")
  })

  it("submits raw text, card type, and card limit to the callback", async () => {
    const user = userEvent.setup()
    const handleSubmit = vi.fn().mockResolvedValue(undefined)

    render(
      <CardImportDialog
        open
        onOpenChange={vi.fn()}
        onSubmit={handleSubmit}
        isPending={false}
      />
    )

    await user.type(
      screen.getByLabelText("卡密内容"),
      "CDK-ABC12345\nCDK-DEF67890"
    )
    await user.selectOptions(screen.getByLabelText("卡片类型"), "es")
    await user.selectOptions(screen.getByLabelText("额度"), "2")
    await user.click(
      within(screen.getByRole("dialog")).getByRole("button", { name: "开始导入" })
    )

    expect(handleSubmit).toHaveBeenCalledWith({
      rawText: "CDK-ABC12345\nCDK-DEF67890",
      cardType: "es",
      cardLimit: 2,
    })
  })

  it("does not offer an all-types option", () => {
    render(
      <CardImportDialog
        open
        onOpenChange={vi.fn()}
        onSubmit={vi.fn().mockResolvedValue(undefined)}
        isPending={false}
      />
    )

    expect(screen.queryByRole("option", { name: "全部" })).not.toBeInTheDocument()
  })
})
