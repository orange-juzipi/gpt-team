import { useEffect, useState } from "react"
import { Check, Copy } from "lucide-react"

import { cn } from "@workspace/ui/lib/utils"

export function CopyableValue({
  value,
  copyValue,
  title,
  className,
  valueClassName,
}: {
  value: string
  copyValue?: string
  title: string
  className?: string
  valueClassName?: string
}) {
  const [copied, setCopied] = useState(false)
  const resolvedCopyValue = copyValue ?? value

  useEffect(() => {
    if (!copied) {
      return undefined
    }

    const timer = window.setTimeout(() => {
      setCopied(false)
    }, 1200)

    return () => {
      window.clearTimeout(timer)
    }
  }, [copied])

  const handleCopy = async () => {
    if (!navigator.clipboard || !resolvedCopyValue) {
      return
    }

    try {
      await navigator.clipboard.writeText(resolvedCopyValue)
      setCopied(true)
    } catch {
      // Best effort clipboard copy in local tooling.
    }
  }

  return (
    <button
      type="button"
      className={cn(
        "group inline-flex max-w-full items-center gap-2 rounded-xl border border-transparent px-2 py-1 text-left transition hover:border-slate-200 hover:bg-slate-50",
        className
      )}
      title={title}
      onClick={() => void handleCopy()}
    >
      <span className={cn("min-w-0 flex-1 truncate", valueClassName)}>{value}</span>
      <span
        className={cn(
          "inline-flex size-6 shrink-0 items-center justify-center rounded-md bg-slate-100 text-slate-400 transition",
          copied
            ? "bg-emerald-100 text-emerald-600 opacity-100"
            : "opacity-0 group-hover:opacity-100"
        )}
        aria-hidden="true"
      >
        {copied ? <Check className="size-3.5" /> : <Copy className="size-3.5" />}
      </span>
    </button>
  )
}
