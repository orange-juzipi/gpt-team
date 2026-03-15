import * as React from "react"

import { cn } from "@workspace/ui/lib/utils"

export function Separator({
  className,
  ...props
}: React.ComponentProps<"hr">) {
  return <hr className={cn("border-border/70", className)} {...props} />
}
