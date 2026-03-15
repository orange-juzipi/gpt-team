import { Badge } from "@workspace/ui/components/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@workspace/ui/components/card"

import { eventTypeLabel, formatDateTime } from "@/lib/format"
import type { CardEventView } from "@/lib/types"

export function CardEventPanel({
  title,
  event,
  children,
}: {
  title: string
  event?: CardEventView
  children?: React.ReactNode
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{title}</CardTitle>
          {event ? (
            <p className="mt-2 text-sm text-muted-foreground">
              最近{eventTypeLabel(event.type)}时间：{formatDateTime(event.createdAt)}
            </p>
          ) : null}
        </div>
        {event?.success ? (
          <Badge variant="success">
            成功
          </Badge>
        ) : null}
      </CardHeader>
      <CardContent className="space-y-3">
        {children ?? (
          <div className="rounded-2xl bg-slate-950 px-4 py-3 font-mono text-xs leading-6 text-slate-100">
            {event?.data ? JSON.stringify(event.data, null, 2) : "暂无数据"}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
