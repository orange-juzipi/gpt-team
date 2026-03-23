import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Plus } from "lucide-react"

import { Button } from "@workspace/ui/components/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@workspace/ui/components/dialog"

import { ConfirmActionDialog } from "@/components/confirm-action-dialog"
import { useMessage } from "@/components/message-context"
import { useFlashMessage } from "@/components/use-flash-message"
import { AccountEmailsDialog } from "@/features/accounts/account-emails-dialog"
import { AccountFormDialog } from "@/features/accounts/account-form-dialog"
import { generateSubAccount } from "@/features/accounts/account-form-helpers"
import { AccountTable } from "@/features/accounts/account-table"
import { useAuth } from "@/features/auth/auth-provider"
import { api } from "@/lib/api"
import type { AccountPayload, AccountRecord } from "@/lib/types"

function buildSubAccountPayload(parent: AccountRecord): AccountPayload {
  return {
    account: generateSubAccount(parent.account),
    password: parent.password,
    type: parent.type,
    startTime: parent.startTime,
    endTime: parent.endTime,
    status: "normal",
    remark: "",
    createMailbox: true,
  }
}

function buildQuickSubAccountPayload(parent: AccountRecord): AccountPayload {
  return {
    account: generateSubAccount(parent.account),
    password: parent.password,
    type: parent.type,
    status: "normal",
    remark: "",
    createMailbox: true,
    useServerTimeSchedule: true,
  }
}

export function SubAccountDialog({
  open,
  onOpenChange,
  account,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  account?: AccountRecord
}) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const message = useMessage()
  const canManage = Boolean(user)
  const [formOpen, setFormOpen] = useState(false)
  const [emailAccount, setEmailAccount] = useState<AccountRecord | undefined>()
  const [editingSubAccount, setEditingSubAccount] = useState<AccountRecord | undefined>()
  const [deletingSubAccount, setDeletingSubAccount] = useState<AccountRecord | undefined>()
  const [createDefaults, setCreateDefaults] = useState<AccountPayload | undefined>()

  const subAccountsQuery = useQuery({
    queryKey: ["subaccounts", account?.id],
    queryFn: () => api.getSubAccounts(account!.id),
    enabled: open && Boolean(account),
  })
  useFlashMessage(subAccountsQuery.isError ? subAccountsQuery.error.message : null)

  const createMutation = useMutation({
    mutationFn: (payload: AccountPayload) =>
      api.createSubAccount(account!.id, payload),
    onSuccess: async () => {
      message.success("子号已创建。")
      setFormOpen(false)
      setEditingSubAccount(undefined)
      setCreateDefaults(undefined)
      await queryClient.invalidateQueries({ queryKey: ["subaccounts", account?.id] })
    },
    onError: (error) => message.error(error.message),
  })

  const updateMutation = useMutation({
    mutationFn: (payload: AccountPayload) =>
      api.updateSubAccount(account!.id, editingSubAccount!.id, payload),
    onSuccess: async () => {
      message.success("子号已更新。")
      setFormOpen(false)
      setEditingSubAccount(undefined)
      setCreateDefaults(undefined)
      await queryClient.invalidateQueries({ queryKey: ["subaccounts", account?.id] })
    },
    onError: (error) => message.error(error.message),
  })

  const deleteMutation = useMutation({
    mutationFn: (subAccountId: number) => api.deleteSubAccount(account!.id, subAccountId),
    onSuccess: async () => {
      message.success("子号已删除。")
      await queryClient.invalidateQueries({ queryKey: ["subaccounts", account?.id] })
    },
    onError: (error) => message.error(error.message),
  })

  const handleDeleteSubAccount = async () => {
    if (!account || !deletingSubAccount) {
      return
    }

    try {
      await deleteMutation.mutateAsync(deletingSubAccount.id)
      setDeletingSubAccount(undefined)
    } catch {
      // Error notice is handled by the mutation callback.
    }
  }

  const handleQuickCreate = async () => {
    if (!account) {
      return
    }

    await createMutation.mutateAsync(buildQuickSubAccountPayload(account))
  }

  if (!account) {
    return null
  }

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-6xl">
          <DialogHeader>
            <DialogTitle>子号管理</DialogTitle>
            <DialogDescription>
              当前主账号：{account.account}。仅 Codex 主账号支持挂载子号。
            </DialogDescription>
          </DialogHeader>
          <DialogBody className="space-y-4">
            {canManage ? (
              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  disabled={createMutation.isPending}
                  onClick={() => void handleQuickCreate()}
                >
                  <Plus className="size-4" />
                  {createMutation.isPending ? "生成中..." : "一键生成子号"}
                </Button>
                <Button
                  onClick={() => {
                    setEditingSubAccount(undefined)
                    setCreateDefaults(buildSubAccountPayload(account))
                    setFormOpen(true)
                  }}
                >
                  <Plus className="size-4" />
                  新增子号
                </Button>
              </div>
            ) : null}
            {subAccountsQuery.isLoading ? (
              <div className="px-2 py-8 text-sm text-muted-foreground">
                子号加载中...
              </div>
            ) : subAccountsQuery.isError ? (
              <div className="px-2 py-8 text-sm text-muted-foreground">
                子号加载失败，请稍后重试。
              </div>
            ) : (
              <AccountTable
                accounts={subAccountsQuery.data ?? []}
                onEdit={
                  canManage
                    ? (item) => {
                        setEditingSubAccount(item)
                        setCreateDefaults(undefined)
                        setFormOpen(true)
                      }
                    : undefined
                }
                onOpenEmails={(item) => setEmailAccount(item)}
                onDelete={
                  canManage
                    ? async (item) => {
                        setDeletingSubAccount(item)
                      }
                    : undefined
                }
                deletingId={deleteMutation.variables}
                readOnly={!canManage}
              />
            )}
          </DialogBody>
        </DialogContent>
      </Dialog>

      {canManage ? (
        <>
          <AccountFormDialog
            open={formOpen}
            onOpenChange={(nextOpen) => {
              setFormOpen(nextOpen)
              if (!nextOpen && !editingSubAccount) {
                setCreateDefaults(undefined)
              }
            }}
            initialValue={editingSubAccount}
            createDefaults={createDefaults}
            title={editingSubAccount ? "编辑子号" : "新增子号"}
            description="默认继承主账号密码与邮箱后缀，你仍然可以按需调整。"
            submitLabel={editingSubAccount ? "保存修改" : "创建子号"}
            isPending={createMutation.isPending || updateMutation.isPending}
            onSubmit={(payload) =>
              editingSubAccount
                ? updateMutation.mutateAsync(payload)
                : createMutation.mutateAsync(payload)
            }
          />

          <ConfirmActionDialog
            open={Boolean(deletingSubAccount)}
            onOpenChange={(nextOpen) => {
              if (!nextOpen) {
                setDeletingSubAccount(undefined)
              }
            }}
            title="删除子号"
            description={
              deletingSubAccount
                ? `确认删除子号 ${deletingSubAccount.account} 吗？`
                : "确认删除这条子号吗？"
            }
            confirmLabel="删除"
            isPending={deleteMutation.isPending}
            onConfirm={handleDeleteSubAccount}
          />

          <AccountEmailsDialog
            open={Boolean(emailAccount)}
            onOpenChange={(nextOpen) => {
              if (!nextOpen) {
                setEmailAccount(undefined)
              }
            }}
            account={emailAccount}
          />
        </>
      ) : null}
    </>
  )
}
