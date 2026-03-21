import type { AccountRecord } from "@/lib/types"

type AccountDeleteTarget = Pick<AccountRecord, "account" | "status" | "type">

export function buildAccountDeleteDescription(account?: AccountDeleteTarget) {
  if (!account) {
    return "确认删除这条账号吗？"
  }

  const relatedAccounts: string[] = []

  if (account.type === "codex") {
    relatedAccounts.push("子号")
  }

  if (account.status === "blocked") {
    relatedAccounts.push("质保号")
  }

  if (relatedAccounts.length === 0) {
    return `确认删除账号 ${account.account} 吗？`
  }

  return `确认删除账号 ${account.account} 吗？若存在${joinWithHe(relatedAccounts)}，将一并删除。`
}

function joinWithHe(items: string[]) {
  if (items.length <= 1) {
    return items[0] ?? ""
  }

  return `${items.slice(0, -1).join("、")}和${items.at(-1)}`
}
