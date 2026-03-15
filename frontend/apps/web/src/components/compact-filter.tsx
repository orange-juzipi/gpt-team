import type { ReactNode } from "react"

import { cn } from "@workspace/ui/lib/utils"

export function CompactFilterSection({
  title,
  summary,
  children,
}: {
  title: string
  summary?: string
  children: ReactNode
}) {
  return (
    <div className="flex flex-col gap-2.5 lg:flex-row lg:items-center lg:justify-between">
      <div className="flex min-w-0 items-center gap-3">
        <div className="text-xs font-semibold tracking-[0.12em] text-slate-400 uppercase">
          {title}
        </div>
        {summary ? (
          <div className="text-sm font-semibold text-slate-500">{summary}</div>
        ) : null}
      </div>
      <div className="flex flex-wrap gap-2">{children}</div>
    </div>
  )
}

export function CompactFilterChip({
  label,
  value,
  active,
  onClick,
}: {
  label: string
  value?: number
  active?: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-sm font-semibold shadow-sm transition",
        active
          ? "border-slate-900 bg-slate-900 text-white"
          : "border-slate-200 bg-white text-slate-600 hover:border-slate-300 hover:text-slate-900"
      )}
    >
      <span>{label}</span>
      {typeof value === "number" ? (
        <span
          className={cn(
            "rounded-full px-2 py-0.5 text-xs font-semibold",
            active ? "bg-white/15 text-white" : "bg-slate-100 text-slate-700"
          )}
        >
          {value}
        </span>
      ) : null}
    </button>
  )
}
