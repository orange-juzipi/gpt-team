import { describe, expect, it } from "vitest"

import {
  generateRandomAccount,
  generateSecurePassword,
} from "@/features/accounts/account-form-helpers"

function createRandomSource(sequence: number[]) {
  let index = 0

  return {
    getRandomValues<T extends ArrayBufferView | null>(array: T): T {
      if (!array) {
        return array
      }

      const view = array as Uint8Array
      for (let offset = 0; offset < view.length; offset += 1) {
        view[offset] = sequence[index % sequence.length] ?? 0
        index += 1
      }

      return array
    },
  }
}

describe("account form helpers", () => {
  it("generates a random account with the selected suffix", () => {
    const account = generateRandomAccount("beta.test", () => 0)

    expect(account).toBe("aaaaa@beta.test")
  })

  it("generates a secure password with mixed character groups", () => {
    const password = generateSecurePassword(18, createRandomSource([0, 1, 2, 3, 4, 5, 6, 7]))

    expect(password).toHaveLength(18)
    expect(password).toMatch(/[a-z]/)
    expect(password).toMatch(/[A-Z]/)
    expect(password).toMatch(/[0-9]/)
    expect(password).toMatch(/[!@#$%^&*\-_=+?]/)
  })

  it("rejects passwords shorter than the required character groups", () => {
    expect(() => generateSecurePassword(3, createRandomSource([0]))).toThrow(
      "密码长度不能小于 4 位。"
    )
  })
})
