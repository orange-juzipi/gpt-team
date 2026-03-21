import { useQuery } from "@tanstack/react-query"
import { RefreshCw } from "lucide-react"

import { Button } from "@workspace/ui/components/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@workspace/ui/components/dialog"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@workspace/ui/components/table"

import { useFlashMessage } from "@/components/use-flash-message"
import { api } from "@/lib/api"
import { formatDateTime } from "@/lib/format"
import type { AccountRecord } from "@/lib/types"

export function AccountEmailsDialog({
  open,
  onOpenChange,
  account,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  account?: AccountRecord
}) {
  const emailsQuery = useQuery({
    queryKey: ["account-emails", account?.id],
    queryFn: () => api.getAccountEmails(account!.id),
    enabled: open && Boolean(account),
  })
  useFlashMessage(emailsQuery.isError ? emailsQuery.error.message : null)

  if (!account) {
    return null
  }

  const items = emailsQuery.data?.items ?? []

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-6xl">
        <DialogHeader className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <DialogTitle>邮件记录</DialogTitle>
            <DialogDescription>
              当前邮箱：{account.account}
            </DialogDescription>
          </div>
          <div className="flex items-center gap-2">
            <div className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-sm text-slate-600">
              共 {items.length} 封
            </div>
            <Button
              variant="outline"
              size="sm"
              disabled={emailsQuery.isFetching}
              onClick={() => void emailsQuery.refetch()}
            >
              <RefreshCw className={`size-4 ${emailsQuery.isFetching ? "animate-spin" : ""}`} />
              {emailsQuery.isFetching ? "刷新中..." : "刷新"}
            </Button>
          </div>
        </DialogHeader>

        <DialogBody className="space-y-4">
          {emailsQuery.isLoading ? (
            <div className="rounded-2xl border border-dashed border-slate-200 px-4 py-10 text-center text-sm text-muted-foreground">
              邮件加载中...
            </div>
          ) : emailsQuery.isError ? (
            <div className="rounded-2xl border border-dashed border-slate-200 px-4 py-10 text-center text-sm text-muted-foreground">
              邮件加载失败，请稍后刷新重试。
            </div>
          ) : items.length === 0 ? (
            <div className="rounded-2xl border border-dashed border-slate-200 px-4 py-10 text-center text-sm text-muted-foreground">
              当前没有查询到邮件记录。
            </div>
          ) : (
            <div className="overflow-x-auto rounded-[24px] border border-slate-200">
              <Table className="min-w-[1040px] table-fixed">
                <TableHeader>
                  <TableRow className="bg-slate-50/70 hover:bg-slate-50/70">
                    <TableHead className="w-[220px] min-w-[220px] max-w-[220px]">发件人</TableHead>
                    <TableHead className="w-[280px] min-w-[280px] max-w-[280px]">主题</TableHead>
                    <TableHead className="w-[360px] min-w-[360px] max-w-[360px]">摘要</TableHead>
                    <TableHead className="w-[200px] min-w-[200px] max-w-[200px]">时间（北京时间）</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((item) => (
                    <TableRow key={`${item.id}-${item.receivedAt}`}>
                      <TableCell className="w-[220px] min-w-[220px] max-w-[220px] align-middle">
                        <div className="space-y-1">
                          <div className="font-medium text-slate-900">
                            {item.fromName || item.from || "未识别"}
                          </div>
                          {item.from && item.fromName && item.from !== item.fromName ? (
                            <div className="text-xs text-muted-foreground">
                              {item.from}
                            </div>
                          ) : null}
                        </div>
                      </TableCell>
                      <TableCell className="w-[280px] min-w-[280px] max-w-[280px] align-middle font-medium text-slate-900">
                        <div className="truncate" title={item.subject || "无主题"}>
                          {item.subject || "无主题"}
                        </div>
                      </TableCell>
                      <TableCell className="w-[360px] min-w-[360px] max-w-[360px] align-middle text-sm leading-6 text-slate-600">
                        <div className="truncate" title={item.preview || "暂无摘要"}>
                          {item.preview || "暂无摘要"}
                        </div>
                      </TableCell>
                      <TableCell className="w-[200px] min-w-[200px] max-w-[200px] align-middle text-sm text-slate-600">
                        {formatDateTime(item.receivedAt, {
                          inputTimeZone: "utc",
                        })}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </DialogBody>
      </DialogContent>
    </Dialog>
  )
}
