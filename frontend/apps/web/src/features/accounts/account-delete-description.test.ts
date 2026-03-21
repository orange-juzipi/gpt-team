import { describe, expect, it } from "vitest"

import { buildAccountDeleteDescription } from "@/features/accounts/account-delete-description"

describe("buildAccountDeleteDescription", () => {
  it("returns a generic description without an account", () => {
    expect(buildAccountDeleteDescription()).toBe("确认删除这条账号吗？")
  })

  it("mentions sub accounts for codex parents", () => {
    expect(
      buildAccountDeleteDescription({
        account: "owner@codex.test",
        type: "codex",
        status: "normal",
      })
    ).toBe("确认删除账号 owner@codex.test 吗？若存在子号，将一并删除。")
  })

  it("mentions warranties for blocked accounts", () => {
    expect(
      buildAccountDeleteDescription({
        account: "owner@plus.test",
        type: "plus",
        status: "blocked",
      })
    ).toBe("确认删除账号 owner@plus.test 吗？若存在质保号，将一并删除。")
  })

  it("mentions both sub accounts and warranties when both can exist", () => {
    expect(
      buildAccountDeleteDescription({
        account: "owner@codex.test",
        type: "codex",
        status: "blocked",
      })
    ).toBe("确认删除账号 owner@codex.test 吗？若存在子号和质保号，将一并删除。")
  })
})
