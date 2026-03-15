import { useEffect, useMemo, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { Download, Plus, RefreshCw, Search, Trash2 } from "lucide-react"

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
import {
  buildCardExportContent,
  buildCardExportFileName,
} from "@/features/cards/card-export"
import { useFlashMessage } from "@/components/use-flash-message"
import { CardImportDialog } from "@/features/cards/card-import-dialog"
import { usePagination } from "@/hooks/use-pagination"
import { api } from "@/lib/api"
import {
  cardLimitLabel,
  cardStatusLabel,
  cardTypeLabel,
  formatDateTime,
} from "@/lib/format"
import type { CardRecord, CardStatus, CardType } from "@/lib/types"

const EMPTY_CARDS: CardRecord[] = []
const CARD_TYPE_FILTERS: CardType[] = ["uk", "us", "es"]
const CARD_STATUS_FILTERS = ["all", "unactivated", "activated"] as const

type CardStatusFilter = "all" | CardStatus

export function CardsPage() {
  const queryClient = useQueryClient()
  const message = useMessage()
  const [importOpen, setImportOpen] = useState(false)
  const [selectedType, setSelectedType] = useState<CardType>(() =>
    resolveCardTypeFromSearch(window.location.search)
  )
  const [selectedStatus, setSelectedStatus] = useState<CardStatusFilter>(() =>
    resolveCardStatusFromSearch(window.location.search)
  )
  const [deletingCard, setDeletingCard] = useState<CardRecord | undefined>()

  const cardsQuery = useQuery({
    queryKey: ["cards"],
    queryFn: api.getCards,
  })

  const importMutation = useMutation({
    mutationFn: api.importCards,
    onSuccess: async (result) => {
      message.success(
        `导入完成：新增 ${result.createdCount} 条，重复或已存在 ${result.duplicates.length} 条。`
      )
      await queryClient.invalidateQueries({ queryKey: ["cards"] })
      setImportOpen(false)
    },
    onError: (error) => {
      message.error(error.message)
    },
  })

  const activateMutation = useMutation({
    mutationFn: api.activateCard,
    onSuccess: async () => {
      message.success("卡密已激活。")
      await queryClient.invalidateQueries({ queryKey: ["cards"] })
    },
    onError: (error) => {
      message.error(error.message)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: api.deleteCard,
    onSuccess: async () => {
      message.success("卡密已删除。")
      await queryClient.invalidateQueries({ queryKey: ["cards"] })
    },
    onError: (error) => {
      message.error(error.message)
    },
  })

  const cards = cardsQuery.data ?? EMPTY_CARDS
  useFlashMessage(cardsQuery.isError ? cardsQuery.error.message : null)

  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search)
    searchParams.set("type", selectedType)

    if (selectedStatus === "all") {
      searchParams.delete("status")
    } else {
      searchParams.set("status", selectedStatus)
    }

    const nextSearch = searchParams.toString()
    const nextURL = `${window.location.pathname}${nextSearch ? `?${nextSearch}` : ""}${window.location.hash}`
    window.history.replaceState(window.history.state, "", nextURL)
  }, [selectedStatus, selectedType])

  const cardsOfSelectedType = useMemo(
    () => cards.filter((item) => item.cardType === selectedType),
    [cards, selectedType]
  )
  const filteredCards = useMemo(
    () =>
      cardsOfSelectedType.filter((item) =>
        selectedStatus === "all" ? true : item.status === selectedStatus
      ),
    [cardsOfSelectedType, selectedStatus]
  )
  const typeCounts = useMemo(
    () =>
      CARD_TYPE_FILTERS.reduce(
        (accumulator, type) => ({
          ...accumulator,
          [type]: cards.filter((item) => item.cardType === type).length,
        }),
        { uk: 0, us: 0, es: 0 } as Record<CardType, number>
      ),
    [cards]
  )
  const stats = useMemo(() => {
    const activated = cardsOfSelectedType.filter(
      (item) => item.status === "activated"
    ).length
    return {
      total: cardsOfSelectedType.length,
      activated,
      unactivated: cardsOfSelectedType.length - activated,
      filtered: filteredCards.length,
    }
  }, [cardsOfSelectedType, filteredCards.length])
  const {
    page,
    setPage,
    pageSize,
    setPageSize,
    totalPages,
    paginatedItems: paginatedCards,
  } = usePagination({
    items: filteredCards,
    resetKeys: [selectedType, selectedStatus],
  })

  const handleDelete = async () => {
    if (!deletingCard) {
      return
    }

    try {
      await deleteMutation.mutateAsync(deletingCard.id)
      setDeletingCard(undefined)
    } catch {
      // Error notice is handled by the mutation callback.
    }
  }

  const handleExport = () => {
    if (filteredCards.length === 0) {
      message.error("当前筛选结果没有可导出的卡密。")
      return
    }

    const content = buildCardExportContent(filteredCards)
    const fileName = buildCardExportFileName(selectedType, filteredCards.length)
    const blob = new Blob([content], { type: "text/plain;charset=utf-8" })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement("a")

    link.href = url
    link.download = fileName
    document.body.append(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)

    message.success(`已导出 ${filteredCards.length} 条卡密。`)
  }

  return (
    <div className="space-y-6">
      <section className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard
          label="当前类型卡密数"
          value={stats.total}
        />
        <MetricCard
          label="已激活"
          value={stats.activated}
        />
        <MetricCard
          label="未激活"
          value={stats.unactivated}
        />
        <MetricCard
          label="当前筛选结果"
          value={stats.filtered}
        />
      </section>

      <section>
        <Card className="overflow-hidden rounded-[28px] border-white/80 bg-white/88 shadow-[0_18px_40px_rgba(148,163,184,0.12)]">
          <CardContent className="space-y-4 p-4 sm:p-5">
            <CompactFilterSection
              title="类型筛选"
              summary={`共 ${typeCounts[selectedType]} 条`}
            >
              {CARD_TYPE_FILTERS.map((type) => (
                <CompactFilterChip
                  key={type}
                  label={cardTypeLabel(type)}
                  value={typeCounts[type]}
                  active={selectedType === type}
                  onClick={() => setSelectedType(type)}
                />
              ))}
            </CompactFilterSection>

            <CompactFilterSection
              title="状态筛选"
              summary={buildStatusFilterSummary(stats.filtered)}
            >
              {CARD_STATUS_FILTERS.map((status) => (
                <CompactFilterChip
                  key={status}
                  label={cardStatusFilterLabel(status)}
                  active={selectedStatus === status}
                  onClick={() => setSelectedStatus(status)}
                />
              ))}
            </CompactFilterSection>
          </CardContent>
        </Card>
      </section>
      <Card>
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <CardTitle>卡密列表</CardTitle>
            <CardDescription>
              按卡片类型筛选展示。点击“查询”进入详情页后，可继续执行账单查询、3DS 查询和全名生日刷新。
            </CardDescription>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Button
              variant="outline"
              disabled={cardsQuery.isFetching}
              onClick={() => void cardsQuery.refetch()}
            >
              <RefreshCw className={`size-4 ${cardsQuery.isFetching ? "animate-spin" : ""}`} />
              {cardsQuery.isFetching ? "刷新中..." : "刷新列表"}
            </Button>
            <Button
              variant="outline"
              disabled={filteredCards.length === 0}
              onClick={handleExport}
            >
              <Download className="size-4" />
              导出文本
            </Button>
            <Button variant="outline" onClick={() => setImportOpen(true)}>
              <Plus className="size-4" />
              批量添加
            </Button>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {cardsQuery.isLoading ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">列表加载中...</div>
          ) : cardsQuery.isError ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              卡密列表加载失败，请稍后重试。
            </div>
          ) : cards.length === 0 ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              还没有任何卡密。先从“批量添加卡密”开始。
            </div>
          ) : filteredCards.length === 0 ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              当前筛选下还没有
              {cardTypeLabel(selectedType)}
              {selectedStatus === "all"
                ? ""
                : `的${cardStatusFilterLabel(selectedStatus)}`}
              。切换筛选或继续导入。
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table className="min-w-[880px] table-fixed">
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[360px] min-w-[360px] max-w-[360px]">卡密</TableHead>
                    <TableHead className="w-[120px] min-w-[120px] max-w-[120px]">额度</TableHead>
                    <TableHead className="w-[140px] min-w-[140px] max-w-[140px]">状态</TableHead>
                    <TableHead className="w-[170px] min-w-[170px] max-w-[170px] text-center">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {paginatedCards.map((card) => (
                    <TableRow key={card.id}>
                      <TableCell className="w-[360px] min-w-[360px] max-w-[360px]">
                        <div className="space-y-1">
                          <CopyableValue
                            value={card.code}
                            title="点击复制卡密"
                            className="max-w-[260px] px-0 py-0 hover:border-transparent hover:bg-transparent"
                            valueClassName="font-mono text-sm font-semibold"
                          />
                          <div className="text-xs text-muted-foreground">
                            最近同步：{card.lastSyncedAt ? formatDateTime(card.lastSyncedAt) : "未同步"}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="w-[120px] min-w-[120px] max-w-[120px]">
                        <Badge variant="outline">{cardLimitLabel(card.cardLimit)}</Badge>
                      </TableCell>
                      <TableCell className="w-[140px] min-w-[140px] max-w-[140px]">
                        <Badge variant={card.status === "activated" ? "success" : "outline"}>
                          {cardStatusLabel(card.status)}
                        </Badge>
                      </TableCell>
                      <TableCell className="w-[170px] min-w-[170px] max-w-[170px]">
                        <div className="flex items-center justify-center gap-2 whitespace-nowrap">
                          <HoverHintButton
                            hint="激活卡密"
                            size="icon-sm"
                            disabled={
                              card.status === "activated" ||
                              (activateMutation.isPending &&
                                activateMutation.variables === card.id)
                            }
                            onClick={() => activateMutation.mutate(card.id)}
                          >
                            <Download className="size-4" />
                          </HoverHintButton>
                          <HoverHintButton
                            hint="查询卡密"
                            size="icon-sm"
                            variant="outline"
                            asChild
                          >
                            <Link to="/cards/$cardId" params={{ cardId: String(card.id) }}>
                              <Search className="size-4" />
                            </Link>
                          </HoverHintButton>
                          <HoverHintButton
                            hint="删除卡密"
                            size="icon-sm"
                            variant="destructive"
                            disabled={
                              deleteMutation.isPending &&
                              deleteMutation.variables === card.id
                            }
                            onClick={() => setDeletingCard(card)}
                          >
                            <Trash2 className="size-4" />
                          </HoverHintButton>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      <PaginationBar
        page={page}
        totalPages={totalPages}
        totalItems={filteredCards.length}
        currentCount={paginatedCards.length}
        pageSize={pageSize}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
      />

      <CardImportDialog
        open={importOpen}
        onOpenChange={setImportOpen}
        onSubmit={(payload) => importMutation.mutateAsync(payload)}
        isPending={importMutation.isPending}
      />

      <ConfirmActionDialog
        open={Boolean(deletingCard)}
        onOpenChange={(open) => {
          if (!open) {
            setDeletingCard(undefined)
          }
        }}
        title="删除卡密"
        description={
          deletingCard
            ? `确认删除卡密 ${deletingCard.code} 吗？`
            : "确认删除这条卡密吗？"
        }
        confirmLabel="删除"
        isPending={deleteMutation.isPending}
        onConfirm={handleDelete}
      />
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

function cardStatusFilterLabel(status: CardStatusFilter) {
  if (status === "all") {
    return "全部状态"
  }

  return cardStatusLabel(status)
}

function buildStatusFilterSummary(count: number) {
  return `匹配 ${count} 条`
}

function resolveCardTypeFromSearch(search: string): CardType {
  const value = new URLSearchParams(search).get("type")
  return CARD_TYPE_FILTERS.includes(value as CardType) ? (value as CardType) : "uk"
}

function resolveCardStatusFromSearch(search: string): CardStatusFilter {
  const value = new URLSearchParams(search).get("status")
  return CARD_STATUS_FILTERS.includes(value as CardStatusFilter)
    ? (value as CardStatusFilter)
    : "all"
}
