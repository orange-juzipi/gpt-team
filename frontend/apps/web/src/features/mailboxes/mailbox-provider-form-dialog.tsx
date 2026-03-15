import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect, useState } from "react"
import { Eye, EyeOff } from "lucide-react"
import { useForm } from "react-hook-form"
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

import type {
  MailboxProviderPayload,
  MailboxProviderRecord,
} from "@/lib/types"

const mailboxProviderSchema = z.object({
  providerType: z.enum(["cloudmail", "duckmail"]),
  domainSuffix: z.string().trim().min(1, "邮箱后缀不能为空"),
  accountEmail: z.string(),
  password: z.string(),
  remark: z.string(),
}).superRefine((value, context) => {
  if (value.providerType === "cloudmail") {
    if (value.accountEmail.trim() === "") {
      context.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["accountEmail"],
        message: "Cloudmail 管理员邮箱不能为空",
      })
    }
    if (value.password.trim() === "") {
      context.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["password"],
        message: "Cloudmail 邮箱密码不能为空",
      })
    }
  }

  if (value.providerType === "duckmail") {
    if (value.password.trim() === "") {
      context.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["password"],
        message: "DuckMail 接口密钥不能为空",
      })
    } else if (!value.password.trim().startsWith("dk_")) {
      context.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["password"],
        message: "DuckMail 接口密钥需以 dk_ 开头",
      })
    }
  }
})

type MailboxProviderFormValues = z.infer<typeof mailboxProviderSchema>

function toFormValues(initial?: MailboxProviderRecord): MailboxProviderFormValues {
  if (!initial) {
    return {
      providerType: "cloudmail",
      domainSuffix: "",
      accountEmail: "",
      password: "",
      remark: "",
    }
  }

  return {
    providerType: initial.providerType,
    domainSuffix: initial.domainSuffix,
    accountEmail: initial.accountEmail,
    password: initial.password,
    remark: initial.remark,
  }
}

export function MailboxProviderFormDialog({
  open,
  onOpenChange,
  initialValue,
  title,
  description,
  submitLabel,
  isPending,
  onSubmit,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialValue?: MailboxProviderRecord
  title: string
  description: string
  submitLabel: string
  isPending: boolean
  onSubmit: (payload: MailboxProviderPayload) => Promise<unknown>
}) {
  const form = useForm<MailboxProviderFormValues>({
    resolver: zodResolver(mailboxProviderSchema),
    defaultValues: toFormValues(initialValue),
  })
  const providerType = form.watch("providerType")
  const [passwordVisible, setPasswordVisible] = useState(false)

  useEffect(() => {
    form.reset(toFormValues(initialValue))
    setPasswordVisible(false)
  }, [form, initialValue, open])

  const handleSubmit = form.handleSubmit(async (values) => {
    await onSubmit({
      providerType: values.providerType,
      domainSuffix: values.domainSuffix,
      accountEmail: values.accountEmail,
      password: values.password,
      remark: values.remark,
    })
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogBody className="space-y-4">
          <FormField
            label="邮局"
            error={form.formState.errors.providerType?.message}
          >
            <Select aria-label="邮局" {...form.register("providerType")}>
              <option value="cloudmail">Cloudmail</option>
              <option value="duckmail">DuckMail</option>
            </Select>
          </FormField>
          <FormField
            label="邮箱后缀"
            error={form.formState.errors.domainSuffix?.message}
          >
            <Input
              aria-label="邮箱后缀"
              placeholder="例如 mail.example"
              {...form.register("domainSuffix")}
            />
          </FormField>
          {providerType === "cloudmail" ? (
            <FormField
              label="管理员邮箱"
              error={form.formState.errors.accountEmail?.message}
            >
              <Input
                aria-label="管理员邮箱"
                placeholder="例如 admin@mail.example"
                {...form.register("accountEmail")}
              />
            </FormField>
          ) : null}
          <FormField
            label={providerType === "duckmail" ? "接口密钥" : "邮箱密码"}
            error={form.formState.errors.password?.message}
          >
            <div className="flex items-center gap-2">
              <Input
                aria-label={providerType === "duckmail" ? "接口密钥" : "邮箱密码"}
                type={passwordVisible ? "text" : "password"}
                placeholder={
                  providerType === "duckmail"
                    ? "例如 dk_xxx"
                    : undefined
                }
                {...form.register("password")}
              />
              <Button
                type="button"
                size="icon-sm"
                variant="outline"
                aria-label={passwordVisible ? "隐藏邮箱密码" : "显示邮箱密码"}
                title={passwordVisible ? "隐藏邮箱密码" : "显示邮箱密码"}
                onClick={() => setPasswordVisible((current) => !current)}
              >
                {passwordVisible ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
              </Button>
            </div>
          </FormField>
          {providerType === "duckmail" ? (
            <p className="text-xs leading-5 text-muted-foreground">
              DuckMail 这里只需要保存接口密钥。取邮件时会直接使用当前账号自己的邮箱和密码换取 token。
            </p>
          ) : null}
          <FormField label="备注">
            <Textarea
              aria-label="备注"
              className="min-h-24"
              {...form.register("remark")}
            />
          </FormField>
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
