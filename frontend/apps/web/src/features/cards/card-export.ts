import { cardStatusLabel, cardTypeLabel } from "@/lib/format"
import type { CardRecord, CardType } from "@/lib/types"

export function buildCardExportContent(cards: CardRecord[]) {
  return cards
    .map((card) =>
      card.status === "activated"
        ? `${card.code}---${cardStatusLabel(card.status)}`
        : card.code
    )
    .join("\n")
}

export function buildCardExportFileName(
  cardType: CardType,
  count: number,
  date = new Date()
) {
  return `${cardTypeLabel(cardType)}-${count}张-${formatExportDate(date)}.txt`
}

function formatExportDate(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, "0")
  const day = String(date.getDate()).padStart(2, "0")
  return `${year}-${month}-${day}`
}
