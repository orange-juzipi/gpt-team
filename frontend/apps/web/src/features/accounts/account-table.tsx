import { useEffect, useState } from "react"
import {
  ArrowDown,
  ArrowUp,
  ArrowUpDown,
  Bot,
  Eye,
  EyeOff,
  Mail,
  Pencil,
  ShieldPlus,
  Trash2,
} from "lucide-react"

import { Badge } from "@workspace/ui/components/badge"
import { Button } from "@workspace/ui/components/button"
import { Input } from "@workspace/ui/components/input"
import { Select } from "@workspace/ui/components/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@workspace/ui/components/table"

import {
  accountStatusLabel,
  accountTypeLabel,
  formatDateTime,
} from "@/lib/format"
import { CopyableValue } from "@/components/copyable-value"
import { HoverHintButton } from "@/components/hover-hint-button"
import type { AccountPayload, AccountRecord, AccountStatus } from "@/lib/types"

export function AccountTable({
  accounts,
  onEdit,
  onDelete,
  onOpenEmails,
  onOpenWarranty,
  onOpenSubAccounts,
  deletingId,
  readOnly,
  onInlineUpdate,
  startTimeSort,
  onToggleStartTimeSort,
}: {
  accounts: AccountRecord[]
  onEdit?: (account: AccountRecord) => void
  onDelete?: (account: AccountRecord) => Promise<void>
  onOpenEmails?: (account: AccountRecord) => void
  onOpenWarranty?: (account: AccountRecord) => void
  onOpenSubAccounts?: (account: AccountRecord) => void
  deletingId?: number
  readOnly?: boolean
  onInlineUpdate?: (
    account: AccountRecord,
    patch: Partial<Pick<AccountPayload, "status" | "remark">>
  ) => Promise<AccountRecord>
  startTimeSort?: "asc" | "desc"
  onToggleStartTimeSort?: () => void
}) {
  const [visiblePasswords, setVisiblePasswords] = useState<number[]>([])

  const toggleVisibility = (accountId: number) => {
    setVisiblePasswords((current) =>
      current.includes(accountId)
        ? current.filter((item) => item !== accountId)
        : [...current, accountId]
    )
  }

  if (accounts.length === 0) {
    return (
      <div className="flex min-h-[220px] items-center justify-center rounded-[28px] border border-dashed border-slate-200/80 px-6 py-10 text-center text-sm text-muted-foreground">
        当前列表为空。
      </div>
    )
  }

  return (
    <div className="overflow-x-auto rounded-[28px]">
      <Table className="min-w-[1450px] table-fixed">
        <TableHeader>
          <TableRow className="bg-slate-50/70 hover:bg-slate-50/70">
            <TableHead className="sticky left-0 z-20 w-[230px] min-w-[230px] max-w-[230px] bg-slate-50/95 py-4 text-center shadow-[8px_0_18px_rgba(15,23,42,0.04)]">
              账号
            </TableHead>
            <TableHead className="w-[240px] min-w-[240px] max-w-[240px] py-4 text-center">密码</TableHead>
            <TableHead className="w-[110px] min-w-[110px] max-w-[110px] py-4 text-center">类型</TableHead>
            <TableHead className="w-[130px] min-w-[130px] max-w-[130px] py-4 text-center">状态</TableHead>
            <TableHead className="w-[170px] min-w-[170px] max-w-[170px] py-4 text-center">
              {onToggleStartTimeSort ? (
                <button
                  type="button"
                  className="inline-flex items-center justify-center gap-1.5 text-xs font-semibold tracking-[0.16em] text-muted-foreground uppercase transition hover:text-slate-900"
                  onClick={onToggleStartTimeSort}
                >
                  开始时间
                  {startTimeSort === "asc" ? (
                    <ArrowUp className="size-3.5" />
                  ) : startTimeSort === "desc" ? (
                    <ArrowDown className="size-3.5" />
                  ) : (
                    <ArrowUpDown className="size-3.5" />
                  )}
                </button>
              ) : (
                "开始时间"
              )}
            </TableHead>
            <TableHead className="w-[170px] min-w-[170px] max-w-[170px] py-4 text-center">结束时间</TableHead>
            <TableHead className="w-[200px] min-w-[200px] max-w-[200px] py-4 text-center">备注</TableHead>
            <TableHead className="sticky right-0 z-20 w-[220px] min-w-[220px] max-w-[220px] bg-slate-50/95 py-4 text-center shadow-[-8px_0_18px_rgba(15,23,42,0.04)]">
              操作
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {accounts.map((account) => {
            const visible = visiblePasswords.includes(account.id)
            return (
              <TableRow key={account.id} className="hover:bg-slate-50/60">
                <TableCell className="sticky left-0 z-10 w-[230px] min-w-[230px] max-w-[230px] bg-white px-4 py-4 align-middle text-center shadow-[8px_0_18px_rgba(15,23,42,0.04)]">
                  <CopyableValue
                    value={account.account}
                    title="点击复制账号"
                    className="mx-auto max-w-[260px] justify-center"
                    valueClassName="font-medium text-slate-900"
                  />
                </TableCell>
                <TableCell className="w-[240px] min-w-[240px] max-w-[240px] px-4 py-4 align-middle text-center">
                  <div className="flex items-center justify-center gap-2">
                    <CopyableValue
                      value={visible ? account.password : account.maskedPassword}
                      copyValue={account.password}
                      title="点击复制密码"
                      className="max-w-[220px] justify-center"
                      valueClassName="font-mono text-xs"
                    />
                    <Button
                      size="icon-sm"
                      variant="ghost"
                      aria-label={visible ? "隐藏密码" : "显示密码"}
                      title={visible ? "隐藏密码" : "显示密码"}
                      onClick={() => toggleVisibility(account.id)}
                    >
                      {visible ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                    </Button>
                  </div>
                </TableCell>
                <TableCell className="w-[110px] min-w-[110px] max-w-[110px] px-4 py-4 align-middle text-center">
                  <Badge
                    variant="outline"
                    className="whitespace-nowrap text-[11px]"
                  >
                    {accountTypeLabel(account.type)}
                  </Badge>
                </TableCell>
                <TableCell className="w-[130px] min-w-[130px] max-w-[130px] px-4 py-4 align-middle text-center">
                  <StatusCell
                    account={account}
                    readOnly={readOnly}
                    onInlineUpdate={onInlineUpdate}
                  />
                </TableCell>
                <TableCell className="w-[170px] min-w-[170px] max-w-[170px] px-4 py-4 align-middle text-center">
                  <DateTimeCell value={account.startTime} />
                </TableCell>
                <TableCell className="w-[170px] min-w-[170px] max-w-[170px] px-4 py-4 align-middle text-center">
                  <DateTimeCell value={account.endTime} />
                </TableCell>
                <TableCell className="w-[200px] min-w-[200px] max-w-[200px] px-4 py-4 align-middle text-center text-sm text-slate-600">
                  <RemarkCell
                    account={account}
                    readOnly={readOnly}
                    onInlineUpdate={onInlineUpdate}
                  />
                </TableCell>
                <TableCell className="sticky right-0 z-10 w-[220px] min-w-[220px] max-w-[220px] bg-white px-4 py-4 align-middle text-center shadow-[-8px_0_18px_rgba(15,23,42,0.04)]">
                  <div className="flex flex-nowrap items-center justify-center gap-1.5">
                    {!readOnly && onEdit ? (
                      <HoverHintButton
                        hint="编辑账号"
                        size="icon-sm"
                        variant="outline"
                        onClick={() => onEdit(account)}
                      >
                        <Pencil className="size-4" />
                      </HoverHintButton>
                    ) : null}
                    {onOpenEmails ? (
                      <HoverHintButton
                        hint="查看邮件"
                        size="icon-sm"
                        variant="outline"
                        onClick={() => onOpenEmails(account)}
                      >
                        <Mail className="size-4" />
                      </HoverHintButton>
                    ) : null}
                    {account.status === "blocked" && onOpenWarranty ? (
                      <HoverHintButton
                        hint="质保账号"
                        size="icon-sm"
                        variant="outline"
                        onClick={() => onOpenWarranty(account)}
                      >
                        <ShieldPlus className="size-4" />
                      </HoverHintButton>
                    ) : null}
                    {account.type === "codex" && onOpenSubAccounts ? (
                      <HoverHintButton
                        hint="子号管理"
                        size="icon-sm"
                        variant="outline"
                        onClick={() => onOpenSubAccounts(account)}
                      >
                        <Bot className="size-4" />
                      </HoverHintButton>
                    ) : null}
                    {!readOnly && onDelete ? (
                      <HoverHintButton
                        hint="删除账号"
                        size="icon-sm"
                        variant="destructive"
                        disabled={deletingId === account.id}
                        onClick={() => void onDelete(account)}
                      >
                        <Trash2 className="size-4" />
                      </HoverHintButton>
                    ) : null}
                  </div>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}

function DateTimeCell({ value }: { value?: string }) {
  const formatted = formatDateTime(value)
  return <div className="whitespace-nowrap text-sm font-medium text-slate-900">{formatted}</div>
}

function StatusCell({
  account,
  readOnly,
  onInlineUpdate,
}: {
  account: AccountRecord
  readOnly?: boolean
  onInlineUpdate?: (
    account: AccountRecord,
    patch: Partial<Pick<AccountPayload, "status" | "remark">>
  ) => Promise<AccountRecord>
}) {
  const [status, setStatus] = useState(account.status)
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    setStatus(account.status)
  }, [account.id, account.status])

  if (readOnly || !onInlineUpdate) {
    return (
      <Badge
        variant={account.status === "blocked" ? "destructive" : "success"}
        className={`whitespace-nowrap text-[11px] ${statusToneClass(account.status)}`}
      >
        {accountStatusLabel(account.status)}
      </Badge>
    )
  }

  const handleChange = async (nextStatus: AccountStatus) => {
    setStatus(nextStatus)
    if (nextStatus === account.status) {
      return
    }

    setIsSaving(true)
    try {
      await onInlineUpdate(account, { status: nextStatus })
    } catch {
      setStatus(account.status)
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <Select
      aria-label={`${account.account} 状态`}
      className={`h-9 rounded-xl px-3 text-center text-xs font-semibold ${statusToneClass(status)}`}
      disabled={isSaving}
      value={status}
      onChange={(event) => void handleChange(event.target.value as AccountStatus)}
    >
      <option value="normal">正常</option>
      <option value="blocked">已封</option>
    </Select>
  )
}

function statusToneClass(status: AccountStatus) {
  return status === "blocked" ? "text-rose-700" : "text-emerald-700"
}

function RemarkCell({
  account,
  readOnly,
  onInlineUpdate,
}: {
  account: AccountRecord
  readOnly?: boolean
  onInlineUpdate?: (
    account: AccountRecord,
    patch: Partial<Pick<AccountPayload, "status" | "remark">>
  ) => Promise<AccountRecord>
}) {
  const [remark, setRemark] = useState(account.remark)
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    setRemark(account.remark)
  }, [account.id, account.remark])

  if (readOnly || !onInlineUpdate) {
    return (
      <div className="mx-auto max-w-[140px] truncate" title={account.remark || "无"}>
        {account.remark || "无"}
      </div>
    )
  }

  const commit = async () => {
    const nextRemark = remark.trim()
    if (nextRemark === account.remark) {
      if (remark !== nextRemark) {
        setRemark(nextRemark)
      }
      return
    }

    setIsSaving(true)
    try {
      await onInlineUpdate(account, { remark: nextRemark })
      setRemark(nextRemark)
    } catch {
      setRemark(account.remark)
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <Input
      aria-label={`${account.account} 备注`}
      className="h-9 rounded-xl px-3 text-center"
      disabled={isSaving}
      maxLength={100}
      placeholder="可直接修改备注"
      value={remark}
      onBlur={() => void commit()}
      onChange={(event) => setRemark(event.target.value)}
      onKeyDown={(event) => {
        if (event.key === "Enter") {
          event.preventDefault()
          void commit()
        }
        if (event.key === "Escape") {
          setRemark(account.remark)
          event.currentTarget.blur()
        }
      }}
    />
  )
}
