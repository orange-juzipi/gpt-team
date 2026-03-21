import { describe, expect, it } from "vitest"

import {
  formatDateTime,
  fromDateTimeLocalValue,
  toDateTimeLocalValue,
} from "@/lib/format"

describe("formatDateTime", () => {
  it("renders ISO timestamps in Beijing time", () => {
    expect(formatDateTime("2026-03-21T04:12:00Z")).toBe("2026年3月21日 12:12")
  })

  it("can treat naive timestamps as UTC before converting to Beijing time", () => {
    expect(
      formatDateTime("2026-03-21 04:12:00", {
        inputTimeZone: "utc",
      })
    ).toBe("2026年3月21日 12:12")
  })
})

describe("Beijing datetime-local conversion", () => {
  it("converts ISO timestamps to Beijing datetime-local values", () => {
    expect(toDateTimeLocalValue("2026-03-21T04:12:00Z")).toBe("2026-03-21T12:12")
  })

  it("converts Beijing datetime-local values back to UTC ISO strings", () => {
    expect(fromDateTimeLocalValue("2026-03-21T12:12")).toBe("2026-03-21T04:12:00.000Z")
  })
})
