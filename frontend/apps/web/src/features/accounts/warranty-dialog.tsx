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
import { useAuth } from "@/features/auth/auth-provider"
import { AccountFormDialog } from "@/features/accounts/account-form-dialog"
import { AccountTable } from "@/features/accounts/account-table"
import { api } from "@/lib/api"
import type { AccountPayload, AccountRecord } from "@/lib/types"

export function WarrantyDialog({
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
  const [editingWarranty, setEditingWarranty] = useState<AccountRecord | undefined>()
  const [deletingWarranty, setDeletingWarranty] = useState<AccountRecord | undefined>()

  const warrantiesQuery = useQuery({
    queryKey: ["warranties", account?.id],
    queryFn: () => api.getWarranties(account!.id),
    enabled: open && Boolean(account),
  })
  useFlashMessage(warrantiesQuery.isError ? warrantiesQuery.error.message : null)

  const createMutation = useMutation({
    mutationFn: (payload: AccountPayload) =>
      api.createWarranty(account!.id, payload),
    onSuccess: async () => {
      message.success("质保账号已创建。")
      setFormOpen(false)
      setEditingWarranty(undefined)
      await queryClient.invalidateQueries({ queryKey: ["warranties", account?.id] })
    },
    onError: (error) => message.error(error.message),
  })

  const updateMutation = useMutation({
    mutationFn: (payload: AccountPayload) =>
      api.updateWarranty(account!.id, editingWarranty!.id, payload),
    onSuccess: async () => {
      message.success("质保账号已更新。")
      setFormOpen(false)
      setEditingWarranty(undefined)
      await queryClient.invalidateQueries({ queryKey: ["warranties", account?.id] })
    },
    onError: (error) => message.error(error.message),
  })

  const deleteMutation = useMutation({
    mutationFn: (warrantyId: number) => api.deleteWarranty(account!.id, warrantyId),
    onSuccess: async () => {
      message.success("质保账号已删除。")
      await queryClient.invalidateQueries({ queryKey: ["warranties", account?.id] })
    },
    onError: (error) => message.error(error.message),
  })

  const handleDeleteWarranty = async () => {
    if (!account || !deletingWarranty) {
      return
    }

    try {
      await deleteMutation.mutateAsync(deletingWarranty.id)
      setDeletingWarranty(undefined)
    } catch {
      // Error notice is handled by the mutation callback.
    }
  }

  if (!account) {
    return null
  }

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-6xl">
          <DialogHeader>
            <DialogTitle>质保账号管理</DialogTitle>
            <DialogDescription>
              当前主账号：{account.account}。只有已封主账号允许挂载质保账号。
            </DialogDescription>
          </DialogHeader>
          <DialogBody className="space-y-4">
            {canManage ? (
              <div className="flex justify-end">
                <Button
                  onClick={() => {
                    setEditingWarranty(undefined)
                    setFormOpen(true)
                  }}
                >
                  <Plus className="size-4" />
                  新增质保账号
                </Button>
              </div>
            ) : null}
            {warrantiesQuery.isLoading ? (
              <div className="px-2 py-8 text-sm text-muted-foreground">
                质保账号加载中...
              </div>
            ) : warrantiesQuery.isError ? (
              <div className="px-2 py-8 text-sm text-muted-foreground">
                质保账号加载失败，请稍后重试。
              </div>
            ) : (
              <AccountTable
                accounts={warrantiesQuery.data ?? []}
                onEdit={
                  canManage
                    ? (item) => {
                        setEditingWarranty(item)
                        setFormOpen(true)
                      }
                    : undefined
                }
                onOpenEmails={(item) => setEmailAccount(item)}
                onDelete={
                  canManage
                    ? async (item) => {
                        setDeletingWarranty(item)
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
            onOpenChange={setFormOpen}
            initialValue={editingWarranty}
            title={editingWarranty ? "编辑质保账号" : "新增质保账号"}
            description="字段结构与主账号一致，但只在当前主账号下展示。"
            submitLabel={editingWarranty ? "保存修改" : "创建质保账号"}
            isPending={createMutation.isPending || updateMutation.isPending}
            onSubmit={(payload) =>
              editingWarranty
                ? updateMutation.mutateAsync(payload)
                : createMutation.mutateAsync(payload)
            }
          />

          <ConfirmActionDialog
            open={Boolean(deletingWarranty)}
            onOpenChange={(open) => {
              if (!open) {
                setDeletingWarranty(undefined)
              }
            }}
            title="删除质保账号"
            description={
              deletingWarranty
                ? `确认删除质保账号 ${deletingWarranty.account} 吗？`
                : "确认删除这条质保账号吗？"
            }
            confirmLabel="删除"
            isPending={deleteMutation.isPending}
            onConfirm={handleDeleteWarranty}
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
