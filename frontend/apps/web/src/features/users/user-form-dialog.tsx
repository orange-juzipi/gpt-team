import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect } from "react"
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

import type { UserPayload, UserRecord } from "@/lib/types"

const userSchema = z.object({
  username: z.string().trim().min(1, "用户名不能为空"),
  password: z.string(),
  role: z.enum(["admin", "member"]),
})

type UserFormValues = z.infer<typeof userSchema>

function toFormValues(initial?: UserRecord): UserFormValues {
  return {
    username: initial?.username ?? "",
    password: "",
    role: initial?.role ?? "member",
  }
}

export function UserFormDialog({
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
  initialValue?: UserRecord
  title: string
  description: string
  submitLabel: string
  isPending: boolean
  onSubmit: (payload: UserPayload) => Promise<unknown>
}) {
  const form = useForm<UserFormValues>({
    resolver: zodResolver(userSchema),
    defaultValues: toFormValues(initialValue),
  })

  useEffect(() => {
    form.reset(toFormValues(initialValue))
  }, [form, initialValue, open])

  const handleSubmit = form.handleSubmit(async (values) => {
    if (!initialValue && !values.password.trim()) {
      form.setError("password", { message: "密码不能为空" })
      return
    }

    await onSubmit({
      username: values.username,
      password: values.password,
      role: values.role,
    })
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogBody className="grid gap-4">
          <FormField label="用户名" error={form.formState.errors.username?.message}>
            <Input aria-label="用户名" {...form.register("username")} />
          </FormField>
          <FormField label="密码" error={form.formState.errors.password?.message}>
            <Input
              aria-label="密码"
              type="password"
              placeholder={initialValue ? "留空表示不修改密码" : ""}
              {...form.register("password")}
            />
          </FormField>
          <FormField label="角色" error={form.formState.errors.role?.message}>
            <Select aria-label="角色" {...form.register("role")}>
              <option value="member">普通用户</option>
              <option value="admin">管理员</option>
            </Select>
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
