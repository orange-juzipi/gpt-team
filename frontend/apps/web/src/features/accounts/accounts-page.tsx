import { type ReactNode, useMemo, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Plus, RefreshCw, UserRound } from "lucide-react"

import { Button } from "@workspace/ui/components/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { cn } from "@workspace/ui/lib/utils"

import { ConfirmActionDialog } from "@/components/confirm-action-dialog"
import { useMessage } from "@/components/message-context"
import { PaginationBar } from "@/components/pagination-bar"
import { AccountAddressDialog } from "@/features/accounts/account-address-dialog"
import { AccountEmailsDialog } from "@/features/accounts/account-emails-dialog"
import { useAuth } from "@/features/auth/auth-provider"
import { useFlashMessage } from "@/components/use-flash-message"
import { AccountFormDialog } from "@/features/accounts/account-form-dialog"
import { SubAccountDialog } from "@/features/accounts/sub-account-dialog"
import { AccountTable } from "@/features/accounts/account-table"
import { WarrantyDialog } from "@/features/accounts/warranty-dialog"
import { usePagination } from "@/hooks/use-pagination"
import { api } from "@/lib/api"
import { accountStatusLabel, accountTypeLabel } from "@/lib/format"
import type {
  AccountPayload,
  AccountRecord,
  AccountStatus,
  AccountType,
} from "@/lib/types"

const EMPTY_ACCOUNTS: AccountRecord[] = []
const ACCOUNT_TYPE_FILTERS = ["all", "business", "plus", "codex"] as const
const ACCOUNT_STATUS_FILTERS = ["all", "normal", "blocked"] as const

type AccountTypeFilter = "all" | AccountType
type AccountStatusFilter = "all" | AccountStatus
type StartTimeSort = "asc" | "desc"

export function AccountsPage() {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const message = useMessage()
  const canManage = Boolean(user)
  const [formOpen, setFormOpen] = useState(false)
  const [addressDialogOpen, setAddressDialogOpen] = useState(false)
  const [editingAccount, setEditingAccount] = useState<AccountRecord | undefined>()
  const [emailAccount, setEmailAccount] = useState<AccountRecord | undefined>()
  const [warrantyParent, setWarrantyParent] = useState<AccountRecord | undefined>()
  const [subAccountParent, setSubAccountParent] = useState<AccountRecord | undefined>()
  const [selectedType, setSelectedType] = useState<AccountTypeFilter>("all")
  const [selectedStatus, setSelectedStatus] = useState<AccountStatusFilter>("all")
  const [startTimeSort, setStartTimeSort] = useState<StartTimeSort>("desc")
  const [deletingAccount, setDeletingAccount] = useState<AccountRecord | undefined>()

  const accountsQuery = useQuery({
    queryKey: ["accounts"],
    queryFn: api.getAccounts,
  })

  const createMutation = useMutation({
    mutationFn: (payload: AccountPayload) => api.createAccount(payload),
    onSuccess: async () => {
      message.success("账号已创建。")
      setFormOpen(false)
      setEditingAccount(undefined)
      await queryClient.invalidateQueries({ queryKey: ["accounts"] })
    },
    onError: (error) => message.error(error.message),
  })

  const updateMutation = useMutation({
    mutationFn: (payload: AccountPayload) =>
      api.updateAccount(editingAccount!.id, payload),
    onSuccess: async () => {
      message.success("账号已更新。")
      setFormOpen(false)
      setEditingAccount(undefined)
      await queryClient.invalidateQueries({ queryKey: ["accounts"] })
    },
    onError: (error) => message.error(error.message),
  })

  const deleteMutation = useMutation({
    mutationFn: (accountId: number) => api.deleteAccount(accountId),
    onSuccess: async () => {
      message.success("账号已删除。")
      await queryClient.invalidateQueries({ queryKey: ["accounts"] })
    },
    onError: (error) => message.error(error.message),
  })

  const accounts = accountsQuery.data ?? EMPTY_ACCOUNTS
  useFlashMessage(accountsQuery.isError ? accountsQuery.error.message : null)

  const filteredAccounts = useMemo(() => {
    const items = accounts.filter((item) => {
      if (selectedType !== "all" && item.type !== selectedType) {
        return false
      }

      if (selectedStatus !== "all" && item.status !== selectedStatus) {
        return false
      }

      return true
    })

    return [...items].sort((left, right) => {
      const leftTime = resolveTimeValue(left.startTime)
      const rightTime = resolveTimeValue(right.startTime)

      if (leftTime === rightTime) {
        return right.id - left.id
      }

      return startTimeSort === "asc" ? leftTime - rightTime : rightTime - leftTime
    })
  }, [accounts, selectedStatus, selectedType, startTimeSort])
  const typeCounts = useMemo(
    () => ({
      all: accounts.length,
      business: accounts.filter((item) => item.type === "business").length,
      plus: accounts.filter((item) => item.type === "plus").length,
      codex: accounts.filter((item) => item.type === "codex").length,
    }),
    [accounts]
  )
  const statusCounts = useMemo(
    () => ({
      all: accounts.length,
      normal: accounts.filter((item) => item.status === "normal").length,
      blocked: accounts.filter((item) => item.status === "blocked").length,
    }),
    [accounts]
  )
  const metrics = useMemo(() => {
    const blockedCount = accounts.filter((item) => item.status === "blocked").length
    const businessCount = accounts.filter((item) => item.type === "business").length
    const plusCount = accounts.filter((item) => item.type === "plus").length
    const codexCount = accounts.filter((item) => item.type === "codex").length

    return {
      total: accounts.length,
      blocked: blockedCount,
      business: businessCount,
      plus: plusCount,
      codex: codexCount,
      normal: accounts.length - blockedCount,
      filtered: filteredAccounts.length,
    }
  }, [accounts, filteredAccounts.length])
  const {
    page,
    setPage,
    pageSize,
    setPageSize,
    totalPages,
    paginatedItems: paginatedAccounts,
  } = usePagination({
    items: filteredAccounts,
    resetKeys: [selectedType, selectedStatus, startTimeSort],
  })

  const handleDeleteAccount = async () => {
    if (!deletingAccount) {
      return
    }

    try {
      await deleteMutation.mutateAsync(deletingAccount.id)
      setDeletingAccount(undefined)
    } catch {
      // Error notice is handled by the mutation callback.
    }
  }

  return (
    <div className="space-y-6">
      <section className="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
        <MetricCard label="账号总数" value={metrics.total} />
        <MetricCard label="Business" value={metrics.business} />
        <MetricCard label="Plus" value={metrics.plus} />
        <MetricCard label="Codex" value={metrics.codex} />
        <MetricCard
          label="当前筛选结果"
          value={metrics.filtered}
        />
      </section>

      <section>
        <Card className="overflow-hidden rounded-[28px] border-white/80 bg-white/88 shadow-[0_18px_40px_rgba(148,163,184,0.12)]">
          <CardContent className="space-y-4 p-4 sm:p-5">
            <CompactFilterRow
              title="类型筛选"
              summary={`共 ${typeCounts[selectedType]} 个`}
            >
              {ACCOUNT_TYPE_FILTERS.map((type) => (
                <FilterChip
                  key={type}
                  label={accountTypeFilterLabel(type)}
                  value={typeCounts[type]}
                  active={selectedType === type}
                  onClick={() => setSelectedType(type)}
                />
              ))}
            </CompactFilterRow>

            <CompactFilterRow
              title="状态筛选"
              summary={`匹配 ${filteredAccounts.length} 个`}
            >
              {ACCOUNT_STATUS_FILTERS.map((status) => (
                <FilterChip
                  key={status}
                  label={accountStatusFilterLabel(status)}
                  value={statusCounts[status]}
                  active={selectedStatus === status}
                  onClick={() => setSelectedStatus(status)}
                />
              ))}
            </CompactFilterRow>
          </CardContent>
        </Card>
      </section>
      <Card>
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <CardTitle>账号列表</CardTitle>
            <CardDescription>
              状态默认“正常”；改为“已封”后才会出现“质保”操作，Codex 账号会出现“子号”操作。
            </CardDescription>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Button
              variant="outline"
              disabled={accountsQuery.isFetching}
              onClick={() => void accountsQuery.refetch()}
            >
              <RefreshCw className={`size-4 ${accountsQuery.isFetching ? "animate-spin" : ""}`} />
              {accountsQuery.isFetching ? "刷新中..." : "刷新列表"}
            </Button>
            <Button
              variant="outline"
              onClick={() => setAddressDialogOpen(true)}
            >
              <UserRound className="size-4" />
              获取地址
            </Button>
            {canManage ? (
              <Button
                variant="outline"
                onClick={() => {
                  setEditingAccount(undefined)
                  setFormOpen(true)
                }}
              >
                <Plus className="size-4" />
                新增账号
              </Button>
            ) : null}
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {accountsQuery.isLoading ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              账号列表加载中...
            </div>
          ) : accountsQuery.isError ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              账号列表加载失败，请稍后重试。
            </div>
          ) : accounts.length === 0 ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              还没有任何账号。先从“新增账号”开始。
            </div>
          ) : filteredAccounts.length === 0 ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              当前筛选下还没有
              {selectedType === "all" ? "" : accountTypeLabel(selectedType)}
              {selectedStatus === "all"
                ? ""
                : `${selectedType === "all" ? "" : "、"}${accountStatusLabel(selectedStatus)}`}
              账号。切换筛选或继续新增。
            </div>
          ) : (
            <AccountTable
              accounts={paginatedAccounts}
              onEdit={
                canManage
                  ? (account) => {
                      setEditingAccount(account)
                      setFormOpen(true)
                    }
                  : undefined
              }
              onOpenEmails={(account) => setEmailAccount(account)}
              onDelete={
                canManage
                  ? async (account) => {
                      setDeletingAccount(account)
                    }
                  : undefined
              }
              onOpenWarranty={(account) => setWarrantyParent(account)}
              onOpenSubAccounts={(account) => setSubAccountParent(account)}
              deletingId={deleteMutation.variables}
              readOnly={!canManage}
              startTimeSort={startTimeSort}
              onToggleStartTimeSort={() =>
                setStartTimeSort((current) => (current === "desc" ? "asc" : "desc"))
              }
            />
          )}
        </CardContent>
      </Card>

      <PaginationBar
        page={page}
        totalPages={totalPages}
        totalItems={filteredAccounts.length}
        currentCount={paginatedAccounts.length}
        pageSize={pageSize}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
      />

      {canManage ? (
        <AccountFormDialog
          open={formOpen}
          onOpenChange={setFormOpen}
          initialValue={editingAccount}
          title={editingAccount ? "编辑账号" : "新增账号"}
          description="账号密码由后端使用 AES-GCM 加密保存，列表默认显示掩码密码。"
          submitLabel={editingAccount ? "保存修改" : "创建账号"}
          isPending={createMutation.isPending || updateMutation.isPending}
          onSubmit={(payload) =>
            editingAccount
              ? updateMutation.mutateAsync(payload)
              : createMutation.mutateAsync(payload)
          }
        />
      ) : null}

      <WarrantyDialog
        open={Boolean(warrantyParent)}
        onOpenChange={(open) => {
          if (!open) {
            setWarrantyParent(undefined)
          }
        }}
        account={warrantyParent}
      />

      <SubAccountDialog
        open={Boolean(subAccountParent)}
        onOpenChange={(open) => {
          if (!open) {
            setSubAccountParent(undefined)
          }
        }}
        account={subAccountParent}
      />

      <AccountEmailsDialog
        open={Boolean(emailAccount)}
        onOpenChange={(open) => {
          if (!open) {
            setEmailAccount(undefined)
          }
        }}
        account={emailAccount}
      />

      <AccountAddressDialog
        open={addressDialogOpen}
        onOpenChange={setAddressDialogOpen}
      />

      {canManage ? (
        <ConfirmActionDialog
          open={Boolean(deletingAccount)}
          onOpenChange={(open) => {
            if (!open) {
              setDeletingAccount(undefined)
            }
          }}
          title="删除账号"
          description={
            deletingAccount
              ? `确认删除账号 ${deletingAccount.account} 吗？`
              : "确认删除这条账号吗？"
          }
          confirmLabel="删除"
          isPending={deleteMutation.isPending}
          onConfirm={handleDeleteAccount}
        />
      ) : null}
    </div>
  )
}

function MetricCard({
  label,
  value,
}: {
  label: string
  value: number | string
}) {
  return (
    <Card className="border-white/70 bg-white/75 shadow-sm">
      <CardHeader className="space-y-1 px-5 py-4">
        <CardDescription className="text-xs">{label}</CardDescription>
        <CardTitle className="text-2xl">{value}</CardTitle>
      </CardHeader>
    </Card>
  )
}

function CompactFilterRow({
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
          <div className="text-sm font-semibold text-slate-500">
            {summary}
          </div>
        ) : null}
      </div>
      <div className="flex flex-wrap gap-2">{children}</div>
    </div>
  )
}

function FilterChip({
  label,
  value,
  active,
  onClick,
}: {
  label: string
  value: number
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
      <span
        className={cn(
          "rounded-full px-2 py-0.5 text-xs font-semibold",
          active ? "bg-white/15 text-white" : "bg-slate-100 text-slate-700"
        )}
      >
        {value}
      </span>
    </button>
  )
}

function accountTypeFilterLabel(type: AccountTypeFilter) {
  if (type === "all") {
    return "全部类型"
  }

  return accountTypeLabel(type)
}

function accountStatusFilterLabel(status: AccountStatusFilter) {
  if (status === "all") {
    return "全部状态"
  }

  return accountStatusLabel(status)
}

function resolveTimeValue(value?: string) {
  if (!value) {
    return Number.MIN_SAFE_INTEGER
  }

  const parsed = new Date(value).getTime()
  if (Number.isNaN(parsed)) {
    return Number.MIN_SAFE_INTEGER
  }

  return parsed
}
