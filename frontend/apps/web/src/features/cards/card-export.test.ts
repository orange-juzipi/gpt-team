import { describe, expect, it } from "vitest"

import { buildCardExportContent, buildCardExportFileName } from "@/features/cards/card-export"
import type { CardRecord } from "@/lib/types"

function createCard(overrides: Partial<CardRecord>): CardRecord {
  return {
    id: 1,
    code: "CDK-DEFAULT",
    cardType: "uk",
    cardLimit: 0,
    status: "unactivated",
    remoteStatus: "",
    lastFour: "",
    expiryDate: "",
    fullName: "",
    birthday: "",
    createdAt: "2026-03-15T00:00:00.000Z",
    updatedAt: "2026-03-15T00:00:00.000Z",
    ...overrides,
  }
}

describe("card export helpers", () => {
  it("only appends status for activated cards", () => {
    const content = buildCardExportContent([
      createCard({ code: "CDK-001", status: "unactivated" }),
      createCard({ id: 2, code: "CDK-002", status: "activated" }),
    ])

    expect(content).toBe("CDK-001\nCDK-002---已激活")
  })

  it("builds the requested file name format", () => {
    expect(buildCardExportFileName("es", 12, new Date("2026-03-15T08:30:00.000Z"))).toBe(
      "西班牙卡-12张-2026-03-15.txt"
    )
  })
})
