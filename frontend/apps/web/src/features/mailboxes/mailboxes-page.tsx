import { useMemo, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Eye, EyeOff, Pencil, Plus, RefreshCw, Trash2 } from "lucide-react"

import { Badge } from "@workspace/ui/components/badge"
import { Button } from "@workspace/ui/components/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@workspace/ui/components/table"

import { CompactFilterChip, CompactFilterSection } from "@/components/compact-filter"
import { ConfirmActionDialog } from "@/components/confirm-action-dialog"
import { CopyableValue } from "@/components/copyable-value"
import { HoverHintButton } from "@/components/hover-hint-button"
import { useMessage } from "@/components/message-context"
import { PaginationBar } from "@/components/pagination-bar"
import { MailboxProviderFormDialog } from "@/features/mailboxes/mailbox-provider-form-dialog"
import { usePagination } from "@/hooks/use-pagination"
import { api } from "@/lib/api"
import { formatDateTime, mailboxProviderTypeLabel } from "@/lib/format"
import type {
  MailboxProviderPayload,
  MailboxProviderRecord,
  MailboxProviderType,
} from "@/lib/types"

const EMPTY_PROVIDERS: MailboxProviderRecord[] = []
const MAILBOX_PROVIDER_FILTERS = ["all", "cloudmail", "duckmail"] as const

type MailboxProviderFilter = "all" | MailboxProviderType

export function MailboxesPage() {
  const queryClient = useQueryClient()
  const message = useMessage()
  const [visiblePasswords, setVisiblePasswords] = useState<number[]>([])
  const [formOpen, setFormOpen] = useState(false)
  const [editingProvider, setEditingProvider] = useState<MailboxProviderRecord | undefined>()
  const [deletingProvider, setDeletingProvider] = useState<MailboxProviderRecord | undefined>()
  const [selectedProviderFilter, setSelectedProviderFilter] =
    useState<MailboxProviderFilter>("all")

  const providersQuery = useQuery({
    queryKey: ["mailbox-providers"],
    queryFn: api.getMailboxProviders,
  })

  const createMutation = useMutation({
    mutationFn: (payload: MailboxProviderPayload) => api.createMailboxProvider(payload),
    onSuccess: async () => {
      message.success("邮箱后缀已创建。")
      setFormOpen(false)
      setEditingProvider(undefined)
      await queryClient.invalidateQueries({ queryKey: ["mailbox-providers"] })
    },
    onError: (error) => message.error(error.message),
  })

  const updateMutation = useMutation({
    mutationFn: (payload: MailboxProviderPayload) =>
      api.updateMailboxProvider(editingProvider!.id, payload),
    onSuccess: async () => {
      message.success("邮箱后缀已更新。")
      setFormOpen(false)
      setEditingProvider(undefined)
      await queryClient.invalidateQueries({ queryKey: ["mailbox-providers"] })
    },
    onError: (error) => message.error(error.message),
  })

  const deleteMutation = useMutation({
    mutationFn: (providerId: number) => api.deleteMailboxProvider(providerId),
    onSuccess: async () => {
      message.success("邮箱后缀已删除。")
      setDeletingProvider(undefined)
      await queryClient.invalidateQueries({ queryKey: ["mailbox-providers"] })
    },
    onError: (error) => message.error(error.message),
  })

  const providers = providersQuery.data ?? EMPTY_PROVIDERS
  const metrics = useMemo(() => {
    return {
      total: providers.length,
      cloudmail: providers.filter((item) => item.providerType === "cloudmail").length,
      duckmail: providers.filter((item) => item.providerType === "duckmail").length,
    }
  }, [providers])
  const providerCounts = useMemo(
    () => ({
      all: providers.length,
      cloudmail: providers.filter((item) => item.providerType === "cloudmail").length,
      duckmail: providers.filter((item) => item.providerType === "duckmail").length,
    }),
    [providers]
  )
  const filteredProviders = useMemo(
    () =>
      providers.filter((item) => {
        if (selectedProviderFilter === "all") {
          return true
        }

        return item.providerType === selectedProviderFilter
      }),
    [providers, selectedProviderFilter]
  )
  const {
    page,
    setPage,
    pageSize,
    setPageSize,
    totalPages,
    paginatedItems: paginatedProviders,
  } = usePagination({
    items: filteredProviders,
    resetKeys: [selectedProviderFilter],
  })

  const toggleVisibility = (providerId: number) => {
    setVisiblePasswords((current) =>
      current.includes(providerId)
        ? current.filter((item) => item !== providerId)
        : [...current, providerId]
    )
  }

  return (
    <div className="space-y-6">
      <section className="grid gap-3 md:grid-cols-3">
        <MetricCard label="邮箱总数" value={metrics.total} />
        <MetricCard label="Cloudmail" value={metrics.cloudmail} />
        <MetricCard label="DuckMail" value={metrics.duckmail} />
      </section>

      <section>
        <Card className="overflow-hidden rounded-[28px] border-white/80 bg-white/88 shadow-[0_18px_40px_rgba(148,163,184,0.12)]">
          <CardContent className="space-y-4 p-4 sm:p-5">
            <CompactFilterSection
              title="邮局筛选"
              summary={`匹配 ${filteredProviders.length} 个`}
            >
              {MAILBOX_PROVIDER_FILTERS.map((filter) => (
                <CompactFilterChip
                  key={filter}
                  label={mailboxProviderFilterLabel(filter)}
                  value={providerCounts[filter]}
                  active={selectedProviderFilter === filter}
                  onClick={() => setSelectedProviderFilter(filter)}
                />
              ))}
            </CompactFilterSection>
          </CardContent>
        </Card>
      </section>

      <Card>
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <CardTitle>邮箱列表</CardTitle>
            <CardDescription>
              维护邮箱后缀与邮局类型；Cloudmail 使用授权账号，DuckMail 只配置接口密钥。
            </CardDescription>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Button
              variant="outline"
              disabled={providersQuery.isFetching}
              onClick={() => void providersQuery.refetch()}
            >
              <RefreshCw className={`size-4 ${providersQuery.isFetching ? "animate-spin" : ""}`} />
              {providersQuery.isFetching ? "刷新中..." : "刷新列表"}
            </Button>
            <Button
              variant="outline"
              onClick={() => {
                setEditingProvider(undefined)
                setFormOpen(true)
              }}
            >
              <Plus className="size-4" />
              新增邮箱
            </Button>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {providersQuery.isLoading ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">邮箱列表加载中...</div>
          ) : providersQuery.isError ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              邮箱列表加载失败，请稍后重试。
            </div>
          ) : providers.length === 0 ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              还没有任何邮箱配置。新增后即可让账号按邮箱后缀切换到对应邮局。
            </div>
          ) : filteredProviders.length === 0 ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              当前筛选下没有匹配的邮箱配置。
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table className="min-w-[1158px] table-fixed">
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[140px] min-w-[140px] max-w-[140px]">邮局</TableHead>
                    <TableHead className="w-[200px] min-w-[200px] max-w-[200px]">邮箱后缀</TableHead>
                    <TableHead className="w-[220px] min-w-[220px] max-w-[220px]">授权账号</TableHead>
                    <TableHead className="w-[192px] min-w-[192px] max-w-[192px]">凭据</TableHead>
                    <TableHead className="w-[140px] min-w-[140px] max-w-[140px]">备注</TableHead>
                    <TableHead className="w-[170px] min-w-[170px] max-w-[170px]">更新时间</TableHead>
                    <TableHead className="sticky right-0 z-20 w-[96px] min-w-[96px] bg-white text-center shadow-[-8px_0_18px_rgba(15,23,42,0.04)]">
                      操作
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {paginatedProviders.map((provider) => {
                    const visible = visiblePasswords.includes(provider.id)
                    return (
                      <TableRow key={provider.id}>
                        <TableCell className="w-[140px] min-w-[140px] max-w-[140px] align-middle">
                          <Badge variant="outline">
                            {mailboxProviderTypeLabel(provider.providerType)}
                          </Badge>
                        </TableCell>
                        <TableCell className="w-[200px] min-w-[200px] max-w-[200px] align-middle">
                          <div className="flex items-center gap-2">
                            <Badge variant="outline">@</Badge>
                            <div className="font-medium">{provider.domainSuffix}</div>
                          </div>
                        </TableCell>
                        <TableCell className="w-[220px] min-w-[220px] max-w-[220px] align-middle">
                          <div className="font-mono text-sm text-slate-700">
                            {provider.providerType === "duckmail"
                              ? "无需授权"
                              : provider.accountEmail || "-"}
                          </div>
                        </TableCell>
                        <TableCell className="align-middle w-[192px] min-w-[192px] max-w-[192px]">
                          {provider.password ? (
                            <div className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-2">
                              <div className="min-w-0 overflow-hidden">
                                <CopyableValue
                                  value={visible ? provider.password : provider.maskedPassword}
                                  copyValue={provider.password}
                                  title={
                                    provider.providerType === "duckmail"
                                      ? "点击复制接口密钥"
                                      : "点击复制邮箱密码"
                                  }
                                  className="w-full max-w-none min-w-0 whitespace-nowrap"
                                  valueClassName="font-mono text-xs"
                                />
                              </div>
                              <Button
                                size="icon-sm"
                                variant="ghost"
                                aria-label={visible ? "隐藏密码" : "显示密码"}
                                title={visible ? "隐藏密码" : "显示密码"}
                                onClick={() => toggleVisibility(provider.id)}
                              >
                                {visible ? (
                                  <EyeOff className="size-4" />
                                ) : (
                                  <Eye className="size-4" />
                                )}
                              </Button>
                            </div>
                          ) : null}
                        </TableCell>
                        <TableCell className="w-[140px] min-w-[140px] max-w-[140px] align-middle text-sm text-muted-foreground">
                          <div className="max-w-[140px] truncate" title={provider.remark || "无"}>
                            {provider.remark || "无"}
                          </div>
                        </TableCell>
                        <TableCell className="w-[170px] min-w-[170px] max-w-[170px] align-middle text-sm text-muted-foreground whitespace-nowrap">
                          {formatDateTime(provider.updatedAt)}
                        </TableCell>
                        <TableCell className="sticky right-0 z-10 align-middle bg-white shadow-[-8px_0_18px_rgba(15,23,42,0.04)]">
                          <div className="flex items-center justify-center gap-2 whitespace-nowrap">
                            <HoverHintButton
                              hint="编辑邮箱"
                              size="icon-sm"
                              variant="outline"
                              onClick={() => {
                                setEditingProvider(provider)
                                setFormOpen(true)
                              }}
                            >
                              <Pencil className="size-4" />
                            </HoverHintButton>
                            <HoverHintButton
                              hint="删除邮箱"
                              size="icon-sm"
                              variant="destructive"
                              disabled={
                                deleteMutation.isPending &&
                                deleteMutation.variables === provider.id
                              }
                              onClick={() => setDeletingProvider(provider)}
                            >
                              <Trash2 className="size-4" />
                            </HoverHintButton>
                          </div>
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      <PaginationBar
        page={page}
        totalPages={totalPages}
        totalItems={filteredProviders.length}
        currentCount={paginatedProviders.length}
        pageSize={pageSize}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
      />

      <MailboxProviderFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        initialValue={editingProvider}
        title={editingProvider ? "编辑邮箱" : "新增邮箱"}
        description="填写邮局类型与邮箱后缀。Cloudmail 配置授权邮箱与密码；DuckMail 只配置接口密钥。"
        submitLabel={editingProvider ? "保存修改" : "创建邮箱"}
        isPending={createMutation.isPending || updateMutation.isPending}
        onSubmit={(payload) =>
          editingProvider
            ? updateMutation.mutateAsync(payload)
            : createMutation.mutateAsync(payload)
        }
      />

      <ConfirmActionDialog
        open={Boolean(deletingProvider)}
        onOpenChange={(open) => {
          if (!open) {
            setDeletingProvider(undefined)
          }
        }}
        title="删除邮箱"
        description={
          deletingProvider
            ? `确认删除邮箱 ${deletingProvider.domainSuffix} 吗？`
            : "确认删除这条邮箱吗？"
        }
        confirmLabel="删除"
        isPending={deleteMutation.isPending}
        onConfirm={async () => {
          if (!deletingProvider) {
            return
          }

          try {
            await deleteMutation.mutateAsync(deletingProvider.id)
          } catch {
            // Error popup is handled by mutation callback.
          }
        }}
      />
    </div>
  )
}

function MetricCard({
  label,
  value,
}: {
  label: string
  value: number
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

function mailboxProviderFilterLabel(filter: MailboxProviderFilter) {
  switch (filter) {
    case "cloudmail":
      return "Cloudmail"
    case "duckmail":
      return "DuckMail"
    default:
      return "全部邮局"
  }
}
