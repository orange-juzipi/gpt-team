import type { ReactNode } from "react"
import { useQuery } from "@tanstack/react-query"
import { CalendarDays, RefreshCw, UserRound } from "lucide-react"

import { Button } from "@workspace/ui/components/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@workspace/ui/components/dialog"

import { CopyableValue } from "@/components/copyable-value"
import { useFlashMessage } from "@/components/use-flash-message"
import { api } from "@/lib/api"

export function AccountAddressDialog({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const profileQuery = useQuery({
    queryKey: ["profiles", "random", open],
    queryFn: api.getRandomProfile,
    enabled: open,
    retry: false,
  })

  useFlashMessage(profileQuery.isError ? profileQuery.error.message : null)

  const profile = profileQuery.data

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>地址资料</DialogTitle>
          <DialogDescription>
            通过美国地址接口获取全名与生日，可直接点击内容复制。
          </DialogDescription>
        </DialogHeader>

        <DialogBody className="space-y-4">
          <div className="flex justify-end">
            <Button
              variant="outline"
              size="sm"
              disabled={profileQuery.isFetching}
              onClick={() => void profileQuery.refetch()}
            >
              <RefreshCw className={`size-4 ${profileQuery.isFetching ? "animate-spin" : ""}`} />
              {profileQuery.isFetching ? "刷新中..." : "刷新"}
            </Button>
          </div>

          {profileQuery.isLoading ? (
            <div className="rounded-[24px] border border-slate-200 bg-slate-50/80 px-4 py-8 text-center text-sm text-muted-foreground">
              资料获取中...
            </div>
          ) : profileQuery.isError ? (
            <div className="rounded-[24px] border border-rose-200 bg-rose-50 px-4 py-8 text-center text-sm text-rose-700">
              获取失败，请重试。
            </div>
          ) : (
            <div className="grid gap-3">
              <AddressInfoCard
                icon={<UserRound className="size-4" />}
                label="全名"
                value={profile?.fullName || "未获取"}
              />
              <AddressInfoCard
                icon={<CalendarDays className="size-4" />}
                label="生日"
                value={profile?.birthday || "未获取"}
              />
            </div>
          )}
        </DialogBody>
      </DialogContent>
    </Dialog>
  )
}

function AddressInfoCard({
  icon,
  label,
  value,
}: {
  icon: ReactNode
  label: string
  value: string
}) {
  return (
    <div className="rounded-[24px] border border-slate-200 bg-slate-50/80 px-4 py-4">
      <div className="mb-3 flex items-center gap-2 text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
        <span className="rounded-full bg-white p-2 text-slate-700 shadow-sm">
          {icon}
        </span>
        {label}
      </div>
      <CopyableValue
        value={value}
        title={`点击复制${label}`}
        className="w-full justify-between rounded-2xl border border-slate-200 bg-white px-3 py-2"
        valueClassName="text-base font-semibold text-slate-900"
      />
    </div>
  )
}
