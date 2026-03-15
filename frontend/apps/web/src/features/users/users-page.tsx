import { useMemo, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Plus, RefreshCw, Trash2, UserPen } from "lucide-react"

import { Button } from "@workspace/ui/components/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Badge } from "@workspace/ui/components/badge"
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
import { HoverHintButton } from "@/components/hover-hint-button"
import { useMessage } from "@/components/message-context"
import { PaginationBar } from "@/components/pagination-bar"
import { useFlashMessage } from "@/components/use-flash-message"
import { useAuth } from "@/features/auth/auth-provider"
import { UserFormDialog } from "@/features/users/user-form-dialog"
import { usePagination } from "@/hooks/use-pagination"
import { api } from "@/lib/api"
import { formatDateTime, userRoleLabel } from "@/lib/format"
import type { UserPayload, UserRecord } from "@/lib/types"

const EMPTY_USERS: UserRecord[] = []
const USER_ROLE_FILTERS = ["all", "admin", "member"] as const

type UserRoleFilter = (typeof USER_ROLE_FILTERS)[number]

export function UsersPage() {
  const { user: currentUser } = useAuth()
  const queryClient = useQueryClient()
  const message = useMessage()
  const [formOpen, setFormOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<UserRecord | undefined>()
  const [deletingUser, setDeletingUser] = useState<UserRecord | undefined>()
  const [selectedRole, setSelectedRole] = useState<UserRoleFilter>("all")

  const usersQuery = useQuery({
    queryKey: ["users"],
    queryFn: api.getUsers,
  })

  const createMutation = useMutation({
    mutationFn: (payload: UserPayload) => api.createUser(payload),
    onSuccess: async () => {
      message.success("用户已创建。")
      setFormOpen(false)
      setEditingUser(undefined)
      await queryClient.invalidateQueries({ queryKey: ["users"] })
    },
    onError: (error) => message.error(error.message),
  })

  const updateMutation = useMutation({
    mutationFn: (payload: UserPayload) => api.updateUser(editingUser!.id, payload),
    onSuccess: async () => {
      message.success("用户已更新。")
      setFormOpen(false)
      setEditingUser(undefined)
      await queryClient.invalidateQueries({ queryKey: ["users"] })
    },
    onError: (error) => message.error(error.message),
  })

  const deleteMutation = useMutation({
    mutationFn: (userId: number) => api.deleteUser(userId),
    onSuccess: async () => {
      message.success("用户已删除。")
      setDeletingUser(undefined)
      await queryClient.invalidateQueries({ queryKey: ["users"] })
    },
    onError: (error) => message.error(error.message),
  })

  const users = usersQuery.data ?? EMPTY_USERS
  useFlashMessage(usersQuery.isError ? usersQuery.error.message : null)

  const metrics = useMemo(() => {
    const adminCount = users.filter((item) => item.role === "admin").length
    return {
      total: users.length,
      admin: adminCount,
      member: users.length - adminCount,
    }
  }, [users])
  const roleCounts = useMemo(
    () => ({
      all: users.length,
      admin: users.filter((item) => item.role === "admin").length,
      member: users.filter((item) => item.role === "member").length,
    }),
    [users]
  )
  const filteredUsers = useMemo(
    () => users.filter((item) => (selectedRole === "all" ? true : item.role === selectedRole)),
    [selectedRole, users]
  )
  const {
    page,
    setPage,
    pageSize,
    setPageSize,
    totalPages,
    paginatedItems: paginatedUsers,
  } = usePagination({
    items: filteredUsers,
    resetKeys: [selectedRole],
  })

  return (
    <div className="space-y-6">
      <section className="grid gap-3 md:grid-cols-3">
        <MetricCard label="用户总数" value={metrics.total} />
        <MetricCard label="管理员" value={metrics.admin} />
        <MetricCard label="普通用户" value={metrics.member} />
      </section>
      <section>
        <Card className="overflow-hidden rounded-[28px] border-white/80 bg-white/88 shadow-[0_18px_40px_rgba(148,163,184,0.12)]">
          <CardContent className="space-y-4 p-4 sm:p-5">
            <CompactFilterSection
              title="角色筛选"
              summary={`匹配 ${filteredUsers.length} 个`}
            >
              {USER_ROLE_FILTERS.map((role) => (
                <CompactFilterChip
                  key={role}
                  label={role === "all" ? "全部角色" : userRoleLabel(role)}
                  value={roleCounts[role]}
                  active={selectedRole === role}
                  onClick={() => setSelectedRole(role)}
                />
              ))}
            </CompactFilterSection>
          </CardContent>
        </Card>
      </section>
      <Card>
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <CardTitle>用户管理</CardTitle>
            <CardDescription>
              管理系统登录账号和角色。管理员可见卡密管理与用户管理，普通用户只可见账号管理。
            </CardDescription>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Button
              variant="outline"
              disabled={usersQuery.isFetching}
              onClick={() => void usersQuery.refetch()}
            >
              <RefreshCw className={`size-4 ${usersQuery.isFetching ? "animate-spin" : ""}`} />
              {usersQuery.isFetching ? "刷新中..." : "刷新列表"}
            </Button>
            <Button
              variant="outline"
              onClick={() => {
                setEditingUser(undefined)
                setFormOpen(true)
              }}
            >
              <Plus className="size-4" />
              新增用户
            </Button>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {usersQuery.isLoading ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">用户列表加载中...</div>
          ) : usersQuery.isError ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              用户列表加载失败，请稍后重试。
            </div>
          ) : users.length === 0 ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              当前还没有用户。系统会在空表时自动创建默认管理员。
            </div>
          ) : filteredUsers.length === 0 ? (
            <div className="px-6 py-12 text-sm text-muted-foreground">
              当前筛选下没有匹配的用户。
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table className="min-w-[900px] table-fixed">
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[320px] min-w-[320px] max-w-[320px]">用户名</TableHead>
                    <TableHead className="w-[140px] min-w-[140px] max-w-[140px]">角色</TableHead>
                    <TableHead className="w-[220px] min-w-[220px] max-w-[220px]">创建时间</TableHead>
                    <TableHead className="w-[140px] min-w-[140px] max-w-[140px] text-center">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {paginatedUsers.map((user) => (
                    <TableRow key={user.id}>
                      <TableCell className="w-[320px] min-w-[320px] max-w-[320px]">
                        <div className="truncate font-medium" title={user.username}>
                          {user.username}
                        </div>
                        {currentUser?.id === user.id ? (
                          <div className="mt-1 text-xs text-muted-foreground">当前登录用户</div>
                        ) : null}
                      </TableCell>
                      <TableCell className="w-[140px] min-w-[140px] max-w-[140px]">
                        <Badge variant={user.role === "admin" ? "success" : "outline"}>
                          {userRoleLabel(user.role)}
                        </Badge>
                      </TableCell>
                      <TableCell className="w-[220px] min-w-[220px] max-w-[220px] text-sm text-muted-foreground">
                        {formatDateTime(user.createdAt)}
                      </TableCell>
                      <TableCell className="w-[140px] min-w-[140px] max-w-[140px]">
                        <div className="flex items-center justify-center gap-2 whitespace-nowrap">
                          <HoverHintButton
                            hint="编辑用户"
                            size="icon-sm"
                            variant="outline"
                            onClick={() => {
                              setEditingUser(user)
                              setFormOpen(true)
                            }}
                          >
                            <UserPen className="size-4" />
                          </HoverHintButton>
                          <HoverHintButton
                            hint="删除用户"
                            size="icon-sm"
                            variant="destructive"
                            disabled={
                              deleteMutation.isPending &&
                              deleteMutation.variables === user.id
                            }
                            onClick={() => setDeletingUser(user)}
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
        totalItems={filteredUsers.length}
        currentCount={paginatedUsers.length}
        pageSize={pageSize}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
      />

      <UserFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        initialValue={editingUser}
        title={editingUser ? "编辑用户" : "新增用户"}
        description={
          editingUser
            ? "留空密码表示保留当前密码不变。"
            : "创建新的后台登录用户。"
        }
        submitLabel={editingUser ? "保存修改" : "创建用户"}
        isPending={createMutation.isPending || updateMutation.isPending}
        onSubmit={(payload) =>
          editingUser
            ? updateMutation.mutateAsync(payload)
            : createMutation.mutateAsync(payload)
        }
      />

      <ConfirmActionDialog
        open={Boolean(deletingUser)}
        onOpenChange={(open) => {
          if (!open) {
            setDeletingUser(undefined)
          }
        }}
        title="删除用户"
        description={
          deletingUser
            ? `确认删除用户 ${deletingUser.username} 吗？`
            : "确认删除这条用户吗？"
        }
        confirmLabel="删除"
        isPending={deleteMutation.isPending}
        onConfirm={async () => {
          if (!deletingUser) {
            return
          }

          try {
            await deleteMutation.mutateAsync(deletingUser.id)
          } catch {
            // Error notice is handled by mutation callback.
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
