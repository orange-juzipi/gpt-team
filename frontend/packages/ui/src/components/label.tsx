import * as React from "react"

import { cn } from "@workspace/ui/lib/utils"

export function Label({
  className,
  ...props
}: React.ComponentProps<"label">) {
  return (
    <label
      className={cn("text-sm font-medium tracking-tight text-foreground", className)}
      {...props}
    />
  )
}
