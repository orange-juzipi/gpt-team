import * as React from "react"

import { cn } from "@workspace/ui/lib/utils"

export function Table({
  className,
  ...props
}: React.ComponentProps<"table">) {
  return <table className={cn("w-full border-separate border-spacing-0", className)} {...props} />
}

export function TableHeader(props: React.ComponentProps<"thead">) {
  return <thead {...props} />
}

export function TableBody(props: React.ComponentProps<"tbody">) {
  return <tbody {...props} />
}

export function TableRow({
  className,
  ...props
}: React.ComponentProps<"tr">) {
  return (
    <tr
      className={cn(
        "transition hover:bg-slate-100/70 [&:not(:last-child)_td]:border-b [&:not(:last-child)_th]:border-b",
        className
      )}
      {...props}
    />
  )
}

export function TableHead({
  className,
  ...props
}: React.ComponentProps<"th">) {
  return (
    <th
      className={cn(
        "border-border/70 px-4 py-3 text-left text-xs font-semibold tracking-[0.16em] text-muted-foreground uppercase",
        className
      )}
      {...props}
    />
  )
}

export function TableCell({
  className,
  ...props
}: React.ComponentProps<"td">) {
  return <td className={cn("border-border/70 px-4 py-3 align-top text-sm", className)} {...props} />
}
