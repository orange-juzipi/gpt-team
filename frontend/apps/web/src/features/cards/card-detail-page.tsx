import { type ReactNode, useEffect, useEffectEvent, useMemo, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Link, useParams } from "@tanstack/react-router"
import {
  ArrowLeft,
  CalendarDays,
  Copy,
  LoaderCircle,
  RefreshCw,
  Search,
  ShieldEllipsis,
  UserRound,
  Wallet,
} from "lucide-react"

import { Badge } from "@workspace/ui/components/badge"
import { Button, buttonVariants } from "@workspace/ui/components/button"
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
import { cn } from "@workspace/ui/lib/utils"

import { useMessage } from "@/components/message-context"
import { useFlashMessage } from "@/components/use-flash-message"
import { CardActionResultDialog } from "@/features/cards/card-action-result-dialog"
import { api } from "@/lib/api"
import {
  cardLimitLabel,
  cardStatusLabel,
  eventTypeLabel,
  cardTypeLabel,
  formatDateTime,
} from "@/lib/format"
import type { CardEventView, CardRecord } from "@/lib/types"

const AUTO_PROFILE_REFRESH_DEDUP_WINDOW_MS = 1000
const AUTO_REMOTE_QUERY_DEDUP_WINDOW_MS = 1000
const DEFAULT_THREE_DS_WINDOW_MINUTES = 30
const autoProfileRefreshRuns = new Map<string, number>()
const autoRemoteQueryRuns = new Map<string, number>()

export function CardDetailPage() {
  const queryClient = useQueryClient()
  const message = useMessage()
  const params = useParams({ strict: false }) as { cardId: string }
  const [billingDialogOpen, setBillingDialogOpen] = useState(false)
  const [threeDSDialogOpen, setThreeDSDialogOpen] = useState(false)

  const detailQuery = useQuery({
    queryKey: ["card", params.cardId],
    queryFn: () => api.getCard(params.cardId),
  })

  const numericCardId = Number(params.cardId)

  const activateMutation = useActionMutation({
    action: () => api.activateCard(numericCardId),
    successMessage: "卡密已激活。",
    queryClient,
    cardId: params.cardId,
    message,
  })
  const queryMutation = useActionMutation({
    action: () => api.queryCard(numericCardId),
    successMessage: "卡信息已刷新。",
    queryClient,
    cardId: params.cardId,
    message,
  })
  const billingMutation = useActionMutation({
    action: () => api.getBilling(numericCardId),
    successMessage: "账单已刷新。",
    queryClient,
    cardId: params.cardId,
    message,
  })
  const profileMutation = useActionMutation({
    action: () => api.refreshProfile(numericCardId),
    successMessage: "全名和生日已刷新。",
    queryClient,
    cardId: params.cardId,
    message,
  })
  const autoQueryMutation = useMutation({
    mutationFn: () => api.queryCard(numericCardId),
    onSuccess: () => {
      void Promise.all([
        queryClient.invalidateQueries({ queryKey: ["cards"] }),
        queryClient.invalidateQueries({ queryKey: ["card", params.cardId] }),
      ])
    },
    onError: (error) => {
      message.error(error.message)
      void Promise.all([
        queryClient.invalidateQueries({ queryKey: ["cards"] }),
        queryClient.invalidateQueries({ queryKey: ["card", params.cardId] }),
      ])
    },
  })
  const autoProfileMutation = useMutation({
    mutationFn: () => api.refreshProfile(numericCardId),
    onSuccess: () => {
      void Promise.all([
        queryClient.invalidateQueries({ queryKey: ["cards"] }),
        queryClient.invalidateQueries({ queryKey: ["card", params.cardId] }),
      ])
    },
    onError: (error) => {
      message.error(error.message)
      void Promise.all([
        queryClient.invalidateQueries({ queryKey: ["cards"] }),
        queryClient.invalidateQueries({ queryKey: ["card", params.cardId] }),
      ])
    },
  })
  const threeDSMutation = useActionMutation({
    action: () => api.getThreeDS(numericCardId, DEFAULT_THREE_DS_WINDOW_MINUTES),
    successMessage: "3DS 验证码已刷新。",
    queryClient,
    cardId: params.cardId,
    message,
  })
  const runAutoCardQuery = useEffectEvent(() => {
    if (!Number.isFinite(numericCardId) || numericCardId <= 0) {
      return
    }

    const now = Date.now()
    const lastRunAt = autoRemoteQueryRuns.get(params.cardId) ?? 0
    if (now - lastRunAt < AUTO_REMOTE_QUERY_DEDUP_WINDOW_MS) {
      return
    }

    autoRemoteQueryRuns.set(params.cardId, now)
    autoQueryMutation.mutate()
  })
  const runAutoProfileRefresh = useEffectEvent(() => {
    if (!Number.isFinite(numericCardId) || numericCardId <= 0) {
      return
    }

    const now = Date.now()
    const lastRunAt = autoProfileRefreshRuns.get(params.cardId) ?? 0
    if (now - lastRunAt < AUTO_PROFILE_REFRESH_DEDUP_WINDOW_MS) {
      return
    }

    autoProfileRefreshRuns.set(params.cardId, now)
    autoProfileMutation.mutate()
  })

  useEffect(() => {
    runAutoCardQuery()
    runAutoProfileRefresh()
  }, [params.cardId])

  const detail = detailQuery.data
  useFlashMessage(
    detailQuery.isError ? detailQuery.error?.message ?? "卡密不存在" : null
  )
  const activationOrQueryEvent = detail?.latestQuery ?? detail?.latestActivation
  const cardSnapshot = useMemo(
    () => extractEnvelope(activationOrQueryEvent),
    [activationOrQueryEvent]
  )
  const billingData = useMemo(
    () => extractEnvelope(detail?.latestBilling)?.data,
    [detail]
  )
  const threeDSData = useMemo(
    () => extractEnvelope(detail?.latestThreeDS)?.data,
    [detail]
  )

  if (detailQuery.isLoading) {
    return <div className="text-sm text-muted-foreground">详情加载中...</div>
  }

  if (detailQuery.isError || !detail) {
    return <div className="text-sm text-muted-foreground">卡密详情加载失败，请返回列表重试。</div>
  }

  const transactions = Array.isArray((billingData as Record<string, unknown>)?.transactions)
    ? ((billingData as Record<string, unknown>).transactions as Array<Record<string, unknown>>)
    : []
  const verifications = Array.isArray((threeDSData as Record<string, unknown>)?.verifications)
    ? ((threeDSData as Record<string, unknown>).verifications as Array<Record<string, unknown>>)
    : []
  const remoteStatusLabel = formatRemoteStatus(
    detail.card.remoteStatus || detail.card.status
  )
  const lastActionLabel = activationOrQueryEvent
    ? eventTypeLabel(activationOrQueryEvent.type)
    : "未执行"
  const snapshotData = cardSnapshot?.data
  const cardNumber = resolveCardNumber(snapshotData)
  const cardNumberDisplay = resolveCardNumberDisplay(snapshotData, detail.card.lastFour)
  const expiryDate = resolveExpiryDate(snapshotData, detail.card.expiryDate)
  const cvv = resolveCVV(snapshotData)
  const fullName = detail.card.fullName || "未获取"
  const isRemoteQuerying = queryMutation.isPending
  const isProfileRefreshing = profileMutation.isPending

  const handleCopyCardSummary = async () => {
    if (!detail || !navigator.clipboard) {
      return
    }

    try {
      await navigator.clipboard.writeText(
        buildCardSnapshotSummary(detail.card, activationOrQueryEvent, cardSnapshot?.data)
      )
      message.success("卡片摘要已复制。")
    } catch {
      message.error("复制失败，请稍后重试。")
    }
  }

  const handleCopyValue = async (value: string, label: string) => {
    if (!navigator.clipboard) {
      return
    }

    if (!value || value === "未获取") {
      message.error(`${label}还未获取。`)
      return
    }

    try {
      await navigator.clipboard.writeText(value)
      message.success(`${label}已复制。`)
    } catch {
      message.error("复制失败，请稍后重试。")
    }
  }

  return (
    <div className="mx-auto max-w-6xl space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-3">
          <Link to="/cards" className={cn(buttonVariants({ variant: "outline" }))}>
            <ArrowLeft className="size-4" />
            返回卡密列表
          </Link>
          <div className="inline-flex items-center gap-2 rounded-full border border-white/75 bg-white/85 px-4 py-2 shadow-sm backdrop-blur">
            <span className="text-[11px] font-semibold tracking-[0.16em] text-muted-foreground uppercase">
              卡密
            </span>
            <span className="font-mono text-sm font-semibold text-slate-900">
              {detail.card.code}
            </span>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <Badge
            variant={detail.card.status === "activated" ? "success" : "outline"}
          >
            {cardStatusLabel(detail.card.status)}
          </Badge>
          <div className="rounded-full border border-white/75 bg-white/85 px-4 py-2 text-sm text-muted-foreground shadow-sm backdrop-blur">
            最近同步：
            <span className="ml-1 font-medium text-foreground">
              {formatDateTime(detail.card.lastSyncedAt)}
            </span>
          </div>
        </div>
      </div>

      <section>
        <Card className="overflow-hidden border-none bg-[linear-gradient(135deg,_rgba(15,23,42,0.98),_rgba(30,64,175,0.95)_58%,_rgba(15,118,110,0.92)_118%)] text-white shadow-[0_24px_48px_rgba(15,23,42,0.2)]">
          <CardHeader className="space-y-4 border-b border-white/10 pb-5">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <CardDescription className="text-slate-200">
                  激活与远程查询
                </CardDescription>
                <CardTitle className="text-2xl font-semibold text-white">
                  操作面板
                </CardTitle>
              </div>
              <div className="rounded-full border border-white/12 bg-white/8 px-3 py-1.5 text-xs text-slate-200">
                {activationOrQueryEvent
                  ? `最近${lastActionLabel} · ${formatDateTime(activationOrQueryEvent.createdAt)}`
                  : "还没有执行过激活或查询"}
              </div>
            </div>

            <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
              <Button
                className="h-11 rounded-2xl bg-white text-slate-900 shadow-none hover:bg-white/95"
                disabled={
                  detail.card.status === "activated" || activateMutation.isPending
                }
                onClick={() => activateMutation.mutate()}
              >
                {activateMutation.isPending ? (
                  <LoaderCircle className="size-4 animate-spin" />
                ) : null}
                {activateMutation.isPending ? "激活中..." : "激活"}
              </Button>
              <Button
                variant="outline"
                className="h-11 rounded-2xl border-white/20 bg-white/8 text-white hover:bg-white/14 hover:text-white"
                disabled={isRemoteQuerying}
                onClick={() => queryMutation.mutate()}
              >
                {isRemoteQuerying ? (
                  <LoaderCircle className="size-4 animate-spin" />
                ) : (
                  <Search className="size-4" />
                )}
                {isRemoteQuerying ? "查询中..." : "查询卡片"}
              </Button>
              <Button
                variant="outline"
                className="h-11 rounded-2xl border-white/20 bg-white/8 text-white hover:bg-white/14 hover:text-white"
                disabled={billingMutation.isPending}
                onClick={() => {
                  setBillingDialogOpen(true)
                  billingMutation.mutate()
                }}
              >
                {billingMutation.isPending ? (
                  <LoaderCircle className="size-4 animate-spin" />
                ) : (
                  <Wallet className="size-4" />
                )}
                {billingMutation.isPending ? "查询中..." : "查询账单"}
              </Button>
              <Button
                variant="outline"
                className="h-11 rounded-2xl border-white/20 bg-white/8 text-white hover:bg-white/14 hover:text-white"
                disabled={threeDSMutation.isPending}
                onClick={() => {
                  setThreeDSDialogOpen(true)
                  threeDSMutation.mutate()
                }}
              >
                {threeDSMutation.isPending ? (
                  <LoaderCircle className="size-4 animate-spin" />
                ) : (
                  <ShieldEllipsis className="size-4" />
                )}
                {threeDSMutation.isPending ? "查询中..." : "查询 3DS"}
              </Button>
            </div>
          </CardHeader>
        </Card>
      </section>

      <section className="grid gap-4 xl:grid-cols-[1.18fr_0.82fr]">
        <div className="space-y-4">
          <CardSnapshotShowcase
            card={detail.card}
            event={activationOrQueryEvent}
            snapshot={snapshotData}
            onCopy={handleCopyCardSummary}
            onCopyValue={handleCopyValue}
          />
        </div>

        <div className="space-y-4">
          <Card className="border-white/70 bg-white/88 shadow-[0_18px_40px_rgba(15,23,42,0.08)] backdrop-blur">
            <CardHeader className="space-y-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <CardDescription>全名与生日</CardDescription>
                  <CardTitle className="text-2xl font-semibold">身份资料</CardTitle>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  className="rounded-full"
                  disabled={isProfileRefreshing}
                  onClick={() => profileMutation.mutate()}
                >
                  {isProfileRefreshing ? (
                    <LoaderCircle className="size-4 animate-spin" />
                  ) : (
                    <RefreshCw className="size-4" />
                  )}
                  {isProfileRefreshing ? "刷新中..." : "刷新"}
                </Button>
              </div>

              <div className="flex flex-wrap items-center gap-2">
                <Badge
                  variant={
                    detail.latestIdentity
                      ? detail.latestIdentity.success
                        ? "success"
                        : "destructive"
                      : "outline"
                  }
                >
                  {detail.latestIdentity
                    ? detail.latestIdentity.success
                      ? "已刷新"
                      : "刷新失败"
                    : "进入页面自动请求"}
                </Badge>
                <p className="text-sm text-muted-foreground">
                  {detail.latestIdentity
                    ? `最近刷新：${formatDateTime(detail.latestIdentity.createdAt)}`
                    : "当前页面每次进入都会重新请求一次。"}
                </p>
              </div>
            </CardHeader>

            <CardContent className="space-y-3">
              <IdentityCard
                icon={<UserRound className="size-4" />}
                label="全名"
                value={fullName}
                copyValue={detail.card.fullName}
                onCopy={handleCopyValue}
              />
              <IdentityCard
                icon={<CalendarDays className="size-4" />}
                label="生日"
                value={detail.card.birthday || "未获取"}
              />
            </CardContent>
          </Card>

          <Card className="border-white/70 bg-white/88 shadow-[0_18px_40px_rgba(15,23,42,0.08)] backdrop-blur">
            <CardHeader className="pb-4">
              <CardDescription>卡片资料</CardDescription>
              <CardTitle className="text-2xl font-semibold">同步信息</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-3 sm:grid-cols-2">
              <CopyableMetricTile
                label="卡号"
                value={cardNumberDisplay}
                copyValue={cardNumber}
                onCopy={handleCopyValue}
              />
              <CopyableMetricTile
                label="有效期"
                value={expiryDate}
                copyValue={expiryDate}
                onCopy={handleCopyValue}
              />
              <CopyableMetricTile
                label="CVV"
                value={cvv}
                copyValue={cvv}
                onCopy={handleCopyValue}
              />
              <MetricTile label="远端状态" value={remoteStatusLabel} />
              <MetricTile
                label="最近动作时间"
                value={
                  activationOrQueryEvent
                    ? formatDateTime(activationOrQueryEvent.createdAt)
                    : "未执行"
                }
              />
            </CardContent>
          </Card>
        </div>
      </section>

      <CardActionResultDialog
        open={billingDialogOpen}
        onOpenChange={setBillingDialogOpen}
        title="账单查询结果"
        description="点击上方操作面板的账单查询后，会在这里展示最新结果。"
        event={detail.latestBilling}
        isPending={billingMutation.isPending}
        onRefresh={() => billingMutation.mutate()}
      >
        {transactions.length > 0 ? (
          <div className="overflow-x-auto rounded-[24px] border border-border/70">
            <Table className="min-w-[840px] table-fixed">
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[320px] min-w-[320px] max-w-[320px]">商户</TableHead>
                  <TableHead className="w-[160px] min-w-[160px] max-w-[160px]">金额</TableHead>
                  <TableHead className="w-[160px] min-w-[160px] max-w-[160px]">状态</TableHead>
                  <TableHead className="w-[200px] min-w-[200px] max-w-[200px]">时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {transactions.map((transaction, index) => (
                  <TableRow key={`${transaction.id ?? index}`}>
                    <TableCell className="w-[320px] min-w-[320px] max-w-[320px]">
                      <div
                        className="truncate"
                        title={String(
                          transaction.merchant ??
                            transaction.merchantName ??
                            "未知商户"
                        )}
                      >
                        {String(
                          transaction.merchant ??
                            transaction.merchantName ??
                            "未知商户"
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="w-[160px] min-w-[160px] max-w-[160px]">
                      {String(transaction.amount ?? "--")} {String(transaction.currency ?? "")}
                    </TableCell>
                    <TableCell className="w-[160px] min-w-[160px] max-w-[160px]">
                      {formatBillingStatus(String(transaction.status ?? ""))}
                    </TableCell>
                    <TableCell className="w-[200px] min-w-[200px] max-w-[200px]">
                      {formatDateTime(
                        String(transaction.createdAt ?? transaction.date ?? "")
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ) : (
          <EmptyActionState
            icon={<Wallet className="size-5" />}
            title="暂无账单明细"
            description="查询成功后，消费记录会展示在这里。"
          />
        )}
      </CardActionResultDialog>

      <CardActionResultDialog
        open={threeDSDialogOpen}
        onOpenChange={setThreeDSDialogOpen}
        title="3DS 验证码查询结果"
        description="点击上方操作面板的 3DS 查询后，会在这里展示最新验证码。"
        event={detail.latestThreeDS}
        isPending={threeDSMutation.isPending}
        onRefresh={() => threeDSMutation.mutate()}
      >
        {verifications.length > 0 ? (
          <div className="space-y-3">
            {verifications.map((verification, index) => (
              <div
                key={`${verification.otp ?? index}`}
                className="rounded-2xl border border-border/70 bg-slate-50 px-4 py-3"
              >
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <div className="font-semibold">
                      验证码 {String(verification.otp ?? "未知")}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {String(verification.merchant ?? "未知商户")} ·{" "}
                      {String(verification.amount ?? "--")}
                    </div>
                  </div>
                  <Badge variant="warning">
                    {formatDateTime(String(verification.receivedAt ?? ""))}
                  </Badge>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <EmptyActionState
            icon={<ShieldEllipsis className="size-5" />}
            title="暂无 3DS 结果"
            description="查询成功后，验证码会展示在这里。"
          />
        )}
      </CardActionResultDialog>
    </div>
  )
}

function useActionMutation({
  action,
  successMessage,
  queryClient,
  cardId,
  message,
}: {
  action: () => Promise<CardEventView>
  successMessage: string
  queryClient: ReturnType<typeof useQueryClient>
  cardId: string
  message: ReturnType<typeof useMessage>
}) {
  return useMutation({
    mutationFn: action,
    onSuccess: async () => {
      message.success(successMessage)
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["cards"] }),
        queryClient.invalidateQueries({ queryKey: ["card", cardId] }),
      ])
    },
    onError: async (error) => {
      message.error(error.message)
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["cards"] }),
        queryClient.invalidateQueries({ queryKey: ["card", cardId] }),
      ])
    },
  })
}

function extractEnvelope(event?: CardEventView) {
  if (!event || typeof event.data !== "object" || event.data === null) {
    return undefined
  }

  return event.data as {
    success?: boolean
    data?: Record<string, unknown>
  }
}

function CardSnapshotShowcase({
  card,
  event,
  snapshot,
  onCopy,
  onCopyValue,
}: {
  card: CardRecord
  event?: CardEventView
  snapshot?: Record<string, unknown>
  onCopy: () => void
  onCopyValue: (value: string, label: string) => void
}) {
  const now = useTicker()
  const code =
    typeof snapshot?.code === "string" && snapshot.code
      ? snapshot.code
      : card.code
  const statusRaw =
    typeof snapshot?.status === "string" && snapshot.status
      ? snapshot.status
      : card.remoteStatus || card.status
  const statusLabel = formatRemoteStatus(statusRaw)
  const cardNumber = resolveCardNumber(snapshot)
  const cardNumberDisplay = resolveCardNumberDisplay(snapshot, card.lastFour)
  const expiryDate = resolveExpiryDate(snapshot, card.expiryDate)
  const preciseExpiryTime = resolvePreciseExpiryTime(snapshot)
  const remainingValidity = formatRemainingValidity(preciseExpiryTime, now)
  const upstreamExpiryDisplay = formatResolvedDateTime(preciseExpiryTime)
  const cvv = resolveCVV(snapshot)
  const remoteCardId =
    typeof snapshot?.cardId === "number"
      ? String(snapshot.cardId)
      : card.remoteCardId
        ? String(card.remoteCardId)
        : "未获取"
  const balance =
    typeof snapshot?.balance === "number"
      ? `$${snapshot.balance.toFixed(2)}`
      : "未提供"
  const remoteCreatedAt =
    typeof snapshot?.createdAt === "string" && snapshot.createdAt
      ? formatDateTime(snapshot.createdAt)
      : "未提供"

  return (
    <div className="space-y-4">
      <div className="overflow-hidden rounded-[28px] border border-slate-200/70 bg-[linear-gradient(180deg,_rgba(247,250,255,0.97),_rgba(234,242,255,0.92)_100%)] p-4 shadow-[0_18px_40px_rgba(30,64,175,0.12)] sm:p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="inline-flex items-center gap-2 rounded-full bg-white/80 px-3 py-1 text-xs font-semibold tracking-[0.16em] text-slate-600 uppercase shadow-sm">
            {event ? eventTypeLabel(event.type) : "卡片快照"}
          </div>
          <div className="text-right text-xs text-slate-400">
            最近同步
            <div className="mt-1 text-sm font-medium text-slate-600">
              {event ? formatDateTime(event.createdAt) : "未执行"}
            </div>
          </div>
        </div>

        <div className="relative mt-4 overflow-hidden rounded-[30px] border border-white/20 bg-[radial-gradient(circle_at_top_left,_rgba(255,255,255,0.24),_transparent_24%),linear-gradient(135deg,_#1f2937_0%,_#1d4ed8_56%,_#0f766e_100%)] px-5 py-5 text-white shadow-[0_24px_40px_rgba(30,64,175,0.2)] sm:px-6">
          <div className="flex items-start justify-between gap-4">
            <div className="rounded-2xl border border-white/15 bg-white/12 p-2.5 shadow-inner">
              <div className="h-8 w-12 rounded-xl border border-white/15 bg-[linear-gradient(135deg,_rgba(255,255,255,0.85),_rgba(255,255,255,0.2))]" />
            </div>
            <div className="flex flex-col items-end gap-3 text-right">
              <div>
                <div className="text-xs font-medium uppercase tracking-[0.18em] text-white/65">
                  {cardTypeLabel(card.cardType)}
                </div>
                <div className="mt-1 text-3xl font-semibold tracking-tight">
                  VIRTUAL
                </div>
              </div>
              <Badge className={statusBadgeClassName(statusRaw)}>{statusLabel}</Badge>
            </div>
          </div>

          <div className="mt-10 space-y-4">
            <button
              type="button"
              className="flex w-full items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/7 px-3 py-2 text-left transition hover:bg-white/12"
              onClick={() => onCopyValue(cardNumber, "卡号")}
            >
              <div className="font-mono text-[1.35rem] tracking-[0.2em] text-white/95 sm:text-[1.6rem]">
                {cardNumberDisplay}
              </div>
              <Copy className="size-4 shrink-0 text-white/70" />
            </button>
            <div className="grid gap-3 sm:grid-cols-3">
              <CardMetaButton
                label="有效期"
                value={expiryDate}
                onClick={() => onCopyValue(expiryDate, "有效期")}
              />
              <CardMetaButton
                label="CVV"
                value={cvv}
                onClick={() => onCopyValue(cvv, "CVV")}
              />
              <CardMetaStat label="额度" value={cardLimitLabel(card.cardLimit)} />
            </div>
          </div>
        </div>

        <div className="mt-4 overflow-hidden rounded-[24px] border border-white/70 bg-white/80 shadow-sm">
          <SnapshotRow
            label="激活码"
            value={code}
            emphasize
            copyValue={code}
            onCopy={onCopyValue}
          />
          <SnapshotRow label="远端卡 ID" value={remoteCardId} />
          <SnapshotRow label="剩余有效时间" value={remainingValidity} />
          <SnapshotRow
            label={preciseExpiryTime ? "上游到期时间" : "上游创建时间"}
            value={preciseExpiryTime ? upstreamExpiryDisplay : remoteCreatedAt}
          />
          <SnapshotRow label="状态" value={statusLabel} />
          <SnapshotRow label="余额" value={balance} />
          <SnapshotRow label="最近同步时间" value={event ? formatDateTime(event.createdAt) : "未执行"} last />
        </div>

        <div className="mt-4 flex justify-end">
          <Button variant="outline" size="sm" onClick={onCopy}>
            <Copy className="size-4" />
            复制卡片摘要
          </Button>
        </div>
      </div>
    </div>
  )
}

function MetricTile({
  label,
  value,
  dark,
}: {
  label: string
  value: string
  dark?: boolean
}) {
  return (
    <div
      className={cn(
        "rounded-2xl border px-4 py-3",
        dark
          ? "border-white/12 bg-white/8"
          : "border-border/70 bg-slate-50"
      )}
    >
      <div
        className={cn(
          "text-xs font-semibold tracking-[0.14em] uppercase",
          dark ? "text-slate-300" : "text-muted-foreground"
        )}
      >
        {label}
      </div>
      <div className={cn("mt-2 text-sm font-medium", dark ? "text-white" : "")}>
        {value}
      </div>
    </div>
  )
}

function EmptyActionState({
  icon,
  title,
  description,
}: {
  icon: ReactNode
  title: string
  description: string
}) {
  return (
    <div className="flex min-h-[220px] flex-col items-center justify-center rounded-[28px] border border-dashed border-border/80 bg-slate-50/80 px-6 py-10 text-center">
      <div className="rounded-full bg-white p-3 text-slate-700 shadow-sm">{icon}</div>
      <div className="mt-4 text-lg font-semibold text-slate-900">{title}</div>
      <p className="mt-2 max-w-md text-sm text-muted-foreground">{description}</p>
    </div>
  )
}

function IdentityCard({
  icon,
  label,
  value,
  copyValue,
  onCopy,
}: {
  icon: ReactNode
  label: string
  value: string
  copyValue?: string
  onCopy?: (value: string, label: string) => void
}) {
  const isCopyable = Boolean(copyValue && onCopy)

  return (
    <button
      type="button"
      className={cn(
        "w-full rounded-[24px] border border-border/70 bg-slate-50/90 px-4 py-4 text-left",
        isCopyable ? "transition hover:border-slate-300 hover:bg-white" : ""
      )}
      onClick={
        isCopyable && onCopy
          ? () => onCopy(copyValue ?? "", label)
          : undefined
      }
    >
      <div className="flex items-center gap-2 text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
        <span className="rounded-full bg-white p-2 text-slate-700 shadow-sm">
          {icon}
        </span>
        {label}
        {isCopyable ? <Copy className="ml-auto size-4 text-slate-400" /> : null}
      </div>
      <div className="mt-4 break-all text-lg font-semibold text-slate-900">
        {value}
      </div>
    </button>
  )
}

function CopyableMetricTile({
  label,
  value,
  copyValue,
  onCopy,
}: {
  label: string
  value: string
  copyValue: string
  onCopy: (value: string, label: string) => void
}) {
  return (
    <button
      type="button"
      className="rounded-2xl border border-border/70 bg-slate-50 px-4 py-3 text-left transition hover:border-slate-300 hover:bg-white"
      onClick={() => onCopy(copyValue, label)}
    >
      <div className="flex items-center justify-between gap-3 text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
        <span>{label}</span>
        <Copy className="size-4" />
      </div>
      <div className="mt-2 break-all text-sm font-medium text-slate-900">
        {value}
      </div>
    </button>
  )
}

function CardMetaButton({
  label,
  value,
  onClick,
}: {
  label: string
  value: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      className="flex min-h-[108px] w-full flex-col rounded-3xl border border-white/10 bg-white/7 px-4 py-3 text-left transition hover:bg-white/12"
      onClick={onClick}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="text-[11px] uppercase tracking-[0.16em] text-white/55">
          {label}
        </div>
        <Copy className="mt-0.5 size-4 shrink-0 text-white/70" />
      </div>
      <div className="mt-auto pt-6 text-[2rem] leading-none font-semibold text-white tabular-nums">
        {value || "未获取"}
      </div>
    </button>
  )
}

function CardMetaStat({
  label,
  value,
}: {
  label: string
  value: string
}) {
  return (
    <div className="flex min-h-[108px] w-full flex-col rounded-3xl border border-white/10 bg-white/7 px-4 py-3 text-left">
      <div className="text-[11px] uppercase tracking-[0.16em] text-white/55">
        {label}
      </div>
      <div className="mt-auto pt-6 text-[2rem] leading-none font-semibold text-white tabular-nums">
        {value || "未获取"}
      </div>
    </div>
  )
}

function SnapshotRow({
  label,
  value,
  emphasize,
  last,
  copyValue,
  onCopy,
}: {
  label: string
  value: string
  emphasize?: boolean
  last?: boolean
  copyValue?: string
  onCopy?: (value: string, label: string) => void
}) {
  const isCopyable = Boolean(copyValue && onCopy)

  return (
    <button
      type="button"
      className={cn(
        "flex w-full items-center justify-between gap-4 px-4 py-4 text-left",
        last ? "" : "border-b border-slate-100",
        isCopyable ? "transition hover:bg-slate-50" : ""
      )}
      onClick={
        isCopyable && onCopy
          ? () => onCopy(copyValue ?? "", label)
          : undefined
      }
    >
      <div className="text-sm text-slate-500">{label}</div>
      <div className="flex items-center gap-3">
        <div
          className={cn(
            "text-right text-sm font-medium text-slate-700",
            emphasize ? "font-mono text-base tracking-[0.14em] text-slate-900" : ""
          )}
        >
          {value}
        </div>
        {isCopyable ? <Copy className="size-4 shrink-0 text-slate-400" /> : null}
      </div>
    </button>
  )
}

function formatMaskedCardNumber(lastFour: string) {
  if (!lastFour) {
    return "••••  ••••  ••••  ••••"
  }

  return `••••  ••••  ••••  ${lastFour}`
}

function formatCardNumberDisplay(cardNumber: string) {
  if (!cardNumber) {
    return "未获取"
  }

  return cardNumber.replace(/\s+/g, "").replace(/(.{4})/g, "$1 ").trim()
}

function resolveCardNumber(snapshot?: Record<string, unknown>) {
  return typeof snapshot?.cardNumber === "string" ? snapshot.cardNumber : ""
}

function resolveCardNumberDisplay(
  snapshot: Record<string, unknown> | undefined,
  lastFourFallback: string
) {
  const cardNumber = resolveCardNumber(snapshot)
  if (cardNumber) {
    return formatCardNumberDisplay(cardNumber)
  }

  return formatMaskedCardNumber(lastFourFallback)
}

function resolveExpiryDate(
  snapshot: Record<string, unknown> | undefined,
  fallback: string
) {
  const monthYearDisplay = resolveExpiryMonthYearDisplay(snapshot)
  const directExpiry = resolveDirectExpiryValue(snapshot)

  if (monthYearDisplay) {
    if (!directExpiry) {
      return monthYearDisplay
    }

    const directDisplay = formatExpiryDisplay(directExpiry)
    if (
      looksLikePreciseDateTimeValue(directExpiry) ||
      (normalizeShortExpiryDisplay(directExpiry) &&
        directDisplay !== monthYearDisplay)
    ) {
      return monthYearDisplay
    }

    return directDisplay
  }

  if (directExpiry) {
    return formatExpiryDisplay(directExpiry)
  }

  if (fallback) {
    return formatExpiryDisplay(fallback)
  }

  return "未获取"
}

function resolveCVV(snapshot?: Record<string, unknown>) {
  if (typeof snapshot?.cvv === "string" && snapshot.cvv) {
    return snapshot.cvv
  }

  return "未获取"
}

function formatRemoteStatus(status: string) {
  switch (status.toLowerCase()) {
    case "active":
    case "activated":
      return "已激活"
    case "used":
      return "已使用"
    case "cancelled":
    case "canceled":
    case "closed":
      return "已销卡"
    case "pending":
      return "处理中"
    case "unactivated":
      return "未激活"
    default:
      return status || "未获取"
  }
}

function formatBillingStatus(status: string) {
  switch (status.trim().toLowerCase()) {
    case "approved":
      return "已通过"
    case "declined":
      return "已拒绝"
    case "pending":
      return "处理中"
    case "settled":
      return "已结算"
    case "reversed":
    case "reversal":
      return "已撤销"
    case "failed":
      return "失败"
    case "success":
    case "succeeded":
      return "成功"
    case "":
      return "--"
    default:
      return status
  }
}

function statusBadgeClassName(status: string) {
  switch (status.toLowerCase()) {
    case "active":
    case "activated":
      return "rounded-full border border-emerald-200/60 bg-emerald-300/15 px-3 py-1 text-xs font-semibold tracking-[0.14em] text-emerald-50 shadow-[0_0_0_1px_rgba(167,243,208,0.18)]"
    case "cancelled":
    case "canceled":
    case "closed":
    case "used":
      return "rounded-full border border-white/45 bg-white/10 px-3 py-1 text-xs font-semibold tracking-[0.14em] text-white shadow-[0_0_0_1px_rgba(255,255,255,0.12)]"
    default:
      return "rounded-full border border-amber-100/55 bg-amber-100/12 px-3 py-1 text-xs font-semibold tracking-[0.14em] text-amber-50 shadow-[0_0_0_1px_rgba(253,230,138,0.12)]"
  }
}

function useTicker() {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = setInterval(() => {
      setNow(Date.now())
    }, 1_000)

    return () => clearInterval(timer)
  }, [])

  return now
}

function resolveDirectExpiryValue(snapshot?: Record<string, unknown>) {
  return pickSnapshotString(snapshot, [
    "expiryDate",
    "expiry",
    "expDate",
    "expireDate",
    "expiry_date",
    "expire_date",
  ])
}

function resolvePreciseExpiryValue(snapshot?: Record<string, unknown>) {
  const preciseValue = pickSnapshotString(snapshot, [
    "expiresAt",
    "expireAt",
    "expiredAt",
    "validUntil",
    "endAt",
    "expiryTime",
  ])
  if (preciseValue) {
    return preciseValue
  }

  const directValue = resolveDirectExpiryValue(snapshot)
  return looksLikePreciseDateTimeValue(directValue) ? directValue : ""
}

function resolveExpiryMonthYearDisplay(snapshot?: Record<string, unknown>) {
  const month = pickSnapshotNumber(snapshot, ["expiryMonth", "expMonth", "month"])
  const year = pickSnapshotNumber(snapshot, ["expiryYear", "expYear", "year"])
  if (!month || !year) {
    return ""
  }

  return `${String(month).padStart(2, "0")}/${String(year).slice(-2)}`
}

function resolvePreciseExpiryTime(snapshot?: Record<string, unknown>) {
  const preciseValue = resolvePreciseExpiryValue(snapshot)
  return preciseValue ? parsePreciseDateValue(preciseValue) : undefined
}

function pickSnapshotString(
  snapshot: Record<string, unknown> | undefined,
  keys: string[]
) {
  if (!snapshot) {
    return ""
  }

  for (const key of keys) {
    const value = snapshot[key]
    if (typeof value === "string" && value.trim()) {
      return value.trim()
    }
  }

  return ""
}

function pickSnapshotNumber(
  snapshot: Record<string, unknown> | undefined,
  keys: string[]
) {
  if (!snapshot) {
    return undefined
  }

  for (const key of keys) {
    const value = snapshot[key]
    if (typeof value === "number" && Number.isFinite(value)) {
      return value
    }
    if (typeof value === "string" && value.trim()) {
      const parsed = Number(value)
      if (Number.isFinite(parsed)) {
        return parsed
      }
    }
  }

  return undefined
}

function formatExpiryDisplay(value: string) {
  const trimmed = value.trim()
  if (!trimmed) {
    return "未获取"
  }

  const shortExpiry = normalizeShortExpiryDisplay(trimmed)
  if (shortExpiry) {
    return shortExpiry
  }

  const expiryDate = looksLikePreciseDateTimeValue(trimmed)
    ? parsePreciseDateValue(trimmed)
    : undefined
  if (expiryDate) {
    return `${String(expiryDate.getMonth() + 1).padStart(2, "0")}/${String(expiryDate.getFullYear()).slice(-2)}`
  }

  return trimmed
}

function parsePreciseDateValue(value: string) {
  if (!looksLikePreciseDateTimeValue(value)) {
    return undefined
  }

  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    return undefined
  }

  return parsed
}

function looksLikePreciseDateTimeValue(value: string) {
  const trimmed = value.trim()
  if (!trimmed) {
    return false
  }

  return (
    /\d{4}[/-]\d{1,2}[/-]\d{1,2}/.test(trimmed) ||
    trimmed.includes("T") ||
    trimmed.includes(":")
  )
}

function normalizeShortExpiryDisplay(value: string) {
  const trimmed = value.trim()
  const monthYearMatch =
    trimmed.match(/^(\d{1,2})[/-](\d{2})$/) ??
    trimmed.match(/^(\d{1,2})[/-](\d{4})$/) ??
    trimmed.match(/^(\d{4})[/-](\d{1,2})$/)

  if (!monthYearMatch) {
    return ""
  }

  let month: number
  let year: number

  if (monthYearMatch[1].length === 4) {
    year = Number(monthYearMatch[1])
    month = Number(monthYearMatch[2])
  } else {
    month = Number(monthYearMatch[1])
    year = Number(monthYearMatch[2])
    if (year < 100) {
      year += 2000
    }
  }

  if (!Number.isFinite(month) || !Number.isFinite(year) || month < 1 || month > 12) {
    return ""
  }

  return `${String(month).padStart(2, "0")}/${String(year).slice(-2)}`
}

function formatRemainingValidity(expiryTime: Date | undefined, now: number) {
  if (!expiryTime) {
    return "未获取"
  }

  const diff = expiryTime.getTime() - now
  if (diff <= 0) {
    return "已到期"
  }

  const totalSeconds = Math.floor(diff / 1_000)
  const totalMinutes = Math.floor(diff / 60_000)
  const totalHours = Math.floor(diff / 3_600_000)
  const totalDays = Math.floor(diff / 86_400_000)

  if (totalDays >= 1) {
    const hours = totalHours % 24
    return hours > 0 ? `${totalDays} 天 ${hours} 小时` : `${totalDays} 天`
  }

  if (totalHours >= 1) {
    const minutes = totalMinutes % 60
    return minutes > 0 ? `${totalHours} 小时 ${minutes} 分钟` : `${totalHours} 小时`
  }

  if (totalMinutes >= 1) {
    const seconds = totalSeconds % 60
    return `${totalMinutes} 分 ${seconds} 秒`
  }

  return `${Math.max(totalSeconds, 1)} 秒`
}

function formatResolvedDateTime(value: Date | undefined) {
  if (!value) {
    return "未获取"
  }

  return formatDateTime(value.toISOString())
}

function buildCardSnapshotSummary(
  card: CardRecord,
  event?: CardEventView,
  snapshot?: Record<string, unknown>
) {
  const cardNumber = resolveCardNumber(snapshot)
  const expiryDate = resolveExpiryDate(snapshot, card.expiryDate)
  const cvv = resolveCVV(snapshot)

  void event

  return [
    cardNumber || formatMaskedCardNumber(card.lastFour),
    cvv,
    expiryDate || "未获取",
  ].join(" ")
}
