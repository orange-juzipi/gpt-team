import type { ReactNode } from "react"

import { cn } from "@workspace/ui/lib/utils"

type NoticeTone = "success" | "error" | "info"

export function Notice({
  tone = "info",
  children,
}: {
  tone?: NoticeTone
  children: ReactNode
}) {
  return (
    <div
      className={cn(
        "rounded-3xl border px-4 py-3 text-sm shadow-sm",
        tone === "success" && "border-emerald-200 bg-emerald-50 text-emerald-800",
        tone === "error" && "border-rose-200 bg-rose-50 text-rose-800",
        tone === "info" && "border-sky-200 bg-sky-50 text-sky-800"
      )}
    >
      {children}
    </div>
  )
}
