import type { ReactNode } from "react"
import { LoaderCircle, RefreshCw } from "lucide-react"

import { Badge } from "@workspace/ui/components/badge"
import { Button } from "@workspace/ui/components/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@workspace/ui/components/dialog"
import { cn } from "@workspace/ui/lib/utils"

import { formatDateTime } from "@/lib/format"
import type { CardEventView } from "@/lib/types"

export function CardActionResultDialog({
  open,
  onOpenChange,
  title,
  description,
  event,
  isPending,
  onRefresh,
  children,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  event?: CardEventView
  isPending?: boolean
  onRefresh: () => void
  children: ReactNode
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl">
        <DialogHeader className="space-y-3">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <DialogTitle>{title}</DialogTitle>
              <DialogDescription>{description}</DialogDescription>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="rounded-full"
              disabled={isPending}
              onClick={onRefresh}
            >
              {isPending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : (
                <RefreshCw className="size-4" />
              )}
              {isPending ? "查询中..." : "重新查询"}
            </Button>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <Badge
              variant={
                event ? (event.success ? "success" : "destructive") : "outline"
              }
            >
              {isPending
                ? "查询中"
                : event
                  ? event.success
                    ? "查询成功"
                    : "查询失败"
                  : "暂无数据"}
            </Badge>
            <span className="text-sm text-muted-foreground">
              {event
                ? `最近查询：${formatDateTime(event.createdAt)}`
                : "点击上方操作按钮后会在这里显示结果。"}
            </span>
          </div>
        </DialogHeader>

        <DialogBody className="space-y-4">
          {event?.errorMessage ? (
            <div className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
              {event.errorMessage}
            </div>
          ) : null}

          <div
            className={cn(
              "space-y-4",
              isPending ? "pointer-events-none opacity-75" : ""
            )}
          >
            {children}
          </div>
        </DialogBody>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            关闭
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
