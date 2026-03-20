import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect, useMemo, useState } from "react"
import { useForm, useWatch } from "react-hook-form"
import { useQuery } from "@tanstack/react-query"
import { z } from "zod"

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
import { Input } from "@workspace/ui/components/input"
import { Label } from "@workspace/ui/components/label"
import { Select } from "@workspace/ui/components/select"
import { Textarea } from "@workspace/ui/components/textarea"

import { useFlashMessage } from "@/components/use-flash-message"
import { api } from "@/lib/api"
import { fromDateTimeLocalValue, toDateTimeLocalValue } from "@/lib/format"
import type { AccountPayload, AccountRecord } from "@/lib/types"
import {
  generateRandomAccount,
  generateSecurePassword,
} from "@/features/accounts/account-form-helpers"

const accountSchema = z.object({
  account: z.string().trim().min(1, "账号不能为空"),
  password: z.string().min(1, "密码不能为空"),
  type: z.enum(["plus", "business", "codex"]),
  startTime: z.string().optional(),
  endTime: z.string().optional(),
  status: z.enum(["normal", "blocked"]),
  remark: z.string(),
  createMailbox: z.boolean(),
})

type AccountFormValues = z.infer<typeof accountSchema>

function buildDefaultSchedule(now = new Date()) {
  const start = new Date(now)
  start.setSeconds(0, 0)

  const end = new Date(start)
  end.setMonth(end.getMonth() + 1)

  return {
    startTime: toDateTimeLocalValue(start.toISOString()),
    endTime: toDateTimeLocalValue(end.toISOString()),
  }
}

function toFormValues(
  initial?: AccountRecord,
  createDefaults?: Partial<AccountPayload>
): AccountFormValues {
  if (!initial) {
    const schedule = buildDefaultSchedule()
    return {
      account: createDefaults?.account ?? "",
      password: createDefaults?.password ?? "",
      type: createDefaults?.type ?? "business",
      startTime: toDateTimeLocalValue(createDefaults?.startTime) || schedule.startTime,
      endTime: toDateTimeLocalValue(createDefaults?.endTime) || schedule.endTime,
      status: createDefaults?.status ?? "normal",
      remark: createDefaults?.remark ?? "",
      createMailbox: createDefaults?.createMailbox ?? true,
    }
  }

  return {
    account: initial.account,
    password: initial.password,
    type: initial.type,
    startTime: toDateTimeLocalValue(initial.startTime),
    endTime: toDateTimeLocalValue(initial.endTime),
    status: initial.status,
    remark: initial.remark,
    createMailbox: true,
  }
}

export function AccountFormDialog({
  open,
  onOpenChange,
  initialValue,
  createDefaults,
  title,
  description,
  submitLabel,
  isPending,
  onSubmit,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialValue?: AccountRecord
  createDefaults?: Partial<AccountPayload>
  title: string
  description: string
  submitLabel: string
  isPending: boolean
  onSubmit: (payload: AccountPayload) => Promise<unknown>
}) {
  const form = useForm<AccountFormValues>({
    resolver: zodResolver(accountSchema),
    defaultValues: toFormValues(initialValue, createDefaults),
  })
  const [selectedRandomDomainSuffix, setSelectedRandomDomainSuffix] = useState("")
  const createMailbox = useWatch({
    control: form.control,
    name: "createMailbox",
  })

  const mailboxProvidersQuery = useQuery({
    queryKey: ["mailbox-providers", "random-suffixes"],
    queryFn: api.getMailboxProviders,
    enabled: open,
    retry: false,
  })
  useFlashMessage(mailboxProvidersQuery.isError ? mailboxProvidersQuery.error.message : null)

  const domainSuffixOptions = useMemo(() => {
    const items = mailboxProvidersQuery.data ?? []
    const uniqueSuffixes = new Set(
      items
        .map((item) => item.domainSuffix.trim())
        .filter((item) => item !== "")
    )

    return [...uniqueSuffixes]
  }, [mailboxProvidersQuery.data])
  const hasMailboxProviders = domainSuffixOptions.length > 0
  const randomDomainSuffix =
    hasMailboxProviders && domainSuffixOptions.includes(selectedRandomDomainSuffix)
      ? selectedRandomDomainSuffix
      : (domainSuffixOptions[0] ?? "")

  useEffect(() => {
    form.reset(toFormValues(initialValue, createDefaults))
  }, [createDefaults, form, initialValue, open])

  const handleSubmit = form.handleSubmit(async (values) => {
    await onSubmit({
      account: values.account,
      password: values.password,
      type: values.type,
      startTime: fromDateTimeLocalValue(values.startTime ?? ""),
      endTime: fromDateTimeLocalValue(values.endTime ?? ""),
      status: values.status,
      remark: values.remark,
      createMailbox: values.createMailbox,
    })
  })

  const handleGenerateAccount = () => {
    if (!randomDomainSuffix) {
      return
    }

    form.setValue("account", generateRandomAccount(randomDomainSuffix), {
      shouldDirty: true,
      shouldTouch: true,
      shouldValidate: true,
    })
  }

  const handleGeneratePassword = () => {
    form.setValue("password", generateSecurePassword(), {
      shouldDirty: true,
      shouldTouch: true,
      shouldValidate: true,
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogBody className="grid gap-4 sm:grid-cols-2">
          <FormField label="账号" error={form.formState.errors.account?.message}>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Input aria-label="账号" {...form.register("account")} />
                <Button
                  type="button"
                  variant="outline"
                  disabled={!hasMailboxProviders}
                  onClick={handleGenerateAccount}
                >
                  随机账号
                </Button>
              </div>
              <div className="flex items-center gap-2">
                <Label className="shrink-0 text-xs text-muted-foreground">随机后缀</Label>
                <Select
                  aria-label="随机邮箱后缀"
                  disabled={!hasMailboxProviders}
                  value={randomDomainSuffix}
                  onChange={(event) => setSelectedRandomDomainSuffix(event.target.value)}
                >
                  {hasMailboxProviders ? (
                    domainSuffixOptions.map((suffix) => (
                      <option key={suffix} value={suffix}>
                        @{suffix}
                      </option>
                    ))
                  ) : (
                    <option value="">暂无可用后缀</option>
                  )}
                </Select>
              </div>
              {hasMailboxProviders ? null : (
                <p className="text-xs text-muted-foreground">
                  尚未创建邮局配置，随机账号暂不可用；你仍然可以手动填写账号并创建。
                </p>
              )}
            </div>
          </FormField>
          <FormField label="密码" error={form.formState.errors.password?.message}>
            <div className="flex items-center gap-2">
              <Input aria-label="密码" {...form.register("password")} />
              {!initialValue ? (
                <Button type="button" variant="outline" onClick={handleGeneratePassword}>
                  随机密码
                </Button>
              ) : null}
            </div>
          </FormField>
          <FormField label="类型" error={form.formState.errors.type?.message}>
            <Select aria-label="类型" {...form.register("type")}>
              <option value="plus">Plus</option>
              <option value="business">Business</option>
              <option value="codex">Codex</option>
            </Select>
          </FormField>
          <FormField label="状态" error={form.formState.errors.status?.message}>
            <Select aria-label="状态" {...form.register("status")}>
              <option value="normal">正常</option>
              <option value="blocked">已封</option>
            </Select>
          </FormField>
          <FormField label="开始时间">
            <Input
              aria-label="开始时间"
              type="datetime-local"
              {...form.register("startTime")}
            />
          </FormField>
          <FormField label="结束时间">
            <Input
              aria-label="结束时间"
              type="datetime-local"
              {...form.register("endTime")}
            />
          </FormField>
          <div className="sm:col-span-2">
            {!initialValue ? (
              <FormField label="邮箱创建">
                <label className="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50/80 px-3 py-3">
                  <Input
                    aria-label="创建邮箱"
                    type="checkbox"
                    className="size-4"
                    checked={createMailbox}
                    onChange={(event) =>
                      form.setValue("createMailbox", event.target.checked, {
                        shouldDirty: true,
                        shouldTouch: true,
                      })
                    }
                  />
                  <div className="space-y-1">
                    <div className="text-sm font-medium text-slate-900">
                      创建账号时同步创建邮箱
                    </div>
                    <div className="text-xs text-muted-foreground">
                      默认开启。关闭后仅保存本地账号，不会向对应邮箱服务创建账户。
                    </div>
                  </div>
                </label>
              </FormField>
            ) : null}
          </div>
          <div className="sm:col-span-2">
            <FormField label="备注">
              <Textarea
                aria-label="备注"
                className="min-h-24"
                {...form.register("remark")}
              />
            </FormField>
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button type="button" disabled={isPending} onClick={() => void handleSubmit()}>
            {isPending ? "提交中..." : submitLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function FormField({
  label,
  children,
  error,
}: {
  label: string
  children: React.ReactNode
  error?: string
}) {
  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      {children}
      {error ? <p className="text-sm text-rose-600">{error}</p> : null}
    </div>
  )
}
