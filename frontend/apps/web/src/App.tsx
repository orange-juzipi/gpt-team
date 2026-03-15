import { useEffect } from "react"
import { Link, Outlet, useLocation, useNavigate } from "@tanstack/react-router"
import {
  AlertTriangle,
  CreditCard,
  LoaderCircle,
  LogOut,
  Mail,
  RefreshCw,
  ShieldCheck,
  Users,
} from "lucide-react"

import { Button } from "@workspace/ui/components/button"

import { cn } from "@workspace/ui/lib/utils"
import { useFlashMessage } from "@/components/use-flash-message"
import { useAuth } from "@/features/auth/auth-provider"
import { LoginPage } from "@/features/auth/login-page"
import { userRoleLabel } from "@/lib/format"

const navigation = [
  {
    to: "/accounts",
    label: "账号管理",
    icon: ShieldCheck,
    adminOnly: false,
  },
  {
    to: "/mailboxes",
    label: "邮箱管理",
    icon: Mail,
    adminOnly: true,
  },
  {
    to: "/cards",
    label: "卡密管理",
    icon: CreditCard,
    adminOnly: true,
  },
  {
    to: "/users",
    label: "用户管理",
    icon: Users,
    adminOnly: true,
  },
] as const

export function App() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, isLoading, error, retry, isRetryPending, logout, isLogoutPending } = useAuth()
  const isCardDetailRoute = /^\/cards\/[^/]+\/?$/.test(location.pathname)
  const showNavigation = !isCardDetailRoute
  const isAdminOnlyRoute =
    location.pathname.startsWith("/cards") ||
    location.pathname.startsWith("/users") ||
    location.pathname.startsWith("/mailboxes")
  const authErrorMessage = error?.message
  const forbiddenMessage =
    user && user.role !== "admin" && isAdminOnlyRoute
      ? "当前用户不是管理员，正在跳转到账号管理。"
      : null

  useFlashMessage(authErrorMessage)
  useFlashMessage(forbiddenMessage, "info")

  useEffect(() => {
    if (isLoading || error) {
      return
    }

    if (!user && location.pathname !== "/login") {
      void navigate({ to: "/login", replace: true })
      return
    }

    if (user && location.pathname === "/login") {
      void navigate({ to: "/accounts", replace: true })
      return
    }

    if (user?.role !== "admin" && isAdminOnlyRoute) {
      void navigate({ to: "/accounts", replace: true })
    }
  }, [error, isAdminOnlyRoute, isLoading, location.pathname, navigate, user])

  if (isLoading) {
    return (
      <div className="flex min-h-svh items-center justify-center bg-[linear-gradient(180deg,_#fffdf6_0%,_#f5f7fb_48%,_#edf2ff_100%)] px-4">
        <div className="text-sm text-muted-foreground">正在检查登录状态...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex min-h-svh items-center justify-center bg-[linear-gradient(180deg,_#fffdf6_0%,_#f5f7fb_48%,_#edf2ff_100%)] px-4">
        <div className="w-full max-w-md rounded-[28px] border border-white/70 bg-white/82 p-6 shadow-[0_18px_48px_rgba(15,23,42,0.12)] backdrop-blur">
          <div className="flex items-start gap-3">
            <div className="rounded-2xl bg-rose-50 p-3 text-rose-600">
              <AlertTriangle className="size-5" />
            </div>
            <div className="space-y-2">
              <div className="text-base font-semibold text-slate-900">系统状态异常</div>
              <div className="text-sm leading-6 text-muted-foreground">
                {error.message || "鉴权初始化失败，请稍后重试。"}
              </div>
            </div>
          </div>
          <div className="mt-5 flex flex-wrap gap-2">
            <Button
              variant="outline"
              disabled={isRetryPending}
              onClick={() => void retry()}
            >
              <RefreshCw className="size-4" />
              重试
            </Button>
            <Button
              variant="outline"
              onClick={() => window.location.reload()}
            >
              刷新页面
            </Button>
          </div>
        </div>
      </div>
    )
  }

  if (!user) {
    return <LoginPage />
  }

  const visibleNavigation = navigation.filter(
    (item) => !item.adminOnly || user.role === "admin"
  )

  if (location.pathname === "/login") {
    return (
      <div className="flex min-h-svh items-center justify-center bg-[linear-gradient(180deg,_#fffdf6_0%,_#f5f7fb_48%,_#edf2ff_100%)] px-4">
        <div className="text-sm text-muted-foreground">正在进入后台...</div>
      </div>
    )
  }

  if (user.role !== "admin" && isAdminOnlyRoute) {
    return (
      <div className="min-h-svh bg-[radial-gradient(circle_at_top_left,_rgba(254,240,138,0.35),_transparent_28%),linear-gradient(180deg,_#fffdf6_0%,_#f5f7fb_48%,_#edf2ff_100%)] text-foreground">
        <div className="mx-auto flex min-h-svh w-full max-w-7xl flex-col px-4 py-4 sm:px-6 sm:py-5 lg:px-8">
          <main className="flex flex-1 items-center justify-center py-5">
            <div className="text-sm text-muted-foreground">正在跳转到账号管理...</div>
          </main>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-svh bg-[radial-gradient(circle_at_top_left,_rgba(254,240,138,0.35),_transparent_28%),linear-gradient(180deg,_#fffdf6_0%,_#f5f7fb_48%,_#edf2ff_100%)] text-foreground">
      <div
        className={cn(
          "mx-auto flex min-h-svh w-full flex-col px-4 sm:px-6 lg:px-8",
          showNavigation
            ? "max-w-7xl py-4 sm:py-5"
            : "max-w-[1240px] py-3 sm:py-4"
        )}
      >
        {showNavigation ? (
          <header className="rounded-[24px] border border-white/70 bg-white/70 px-4 py-3 shadow-[0_14px_44px_rgba(15,23,42,0.07)] backdrop-blur sm:px-5">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div className="flex flex-wrap items-center gap-2 lg:flex-1">
                {visibleNavigation.map((item) => {
                  const Icon = item.icon
                  const active = location.pathname.startsWith(item.to)

                  return (
                    <Link
                      key={item.to}
                      to={item.to}
                      className={cn(
                        "min-w-[180px] rounded-[20px] border px-4 py-3 transition lg:min-w-0",
                        active
                          ? "border-slate-900 bg-slate-900 text-white shadow-lg"
                          : "border-white/70 bg-white/70 text-foreground hover:border-slate-300 hover:bg-white"
                      )}
                    >
                      <div className="flex items-center gap-3">
                        <div
                          className={cn(
                            "rounded-xl p-2.5",
                            active
                              ? "bg-white/12 text-white"
                              : "bg-slate-100 text-slate-700"
                          )}
                        >
                          <Icon className="size-4.5" />
                        </div>
                        <div className="text-base font-semibold">{item.label}</div>
                      </div>
                    </Link>
                  )
                })}
              </div>

              <div className="flex items-center justify-between gap-3 rounded-[24px] border border-white/75 bg-[linear-gradient(135deg,_rgba(255,255,255,0.92),_rgba(248,250,252,0.88))] p-2.5 shadow-[0_12px_30px_rgba(15,23,42,0.06)] lg:min-w-[340px] lg:shrink-0">
                <div className="flex min-w-0 items-center gap-3">
                  <div className="flex size-12 shrink-0 items-center justify-center rounded-[18px] bg-[linear-gradient(135deg,_#0f172a,_#334155)] text-sm font-semibold text-white shadow-[0_10px_24px_rgba(15,23,42,0.18)]">
                    {resolveUserMonogram(user.username)}
                  </div>
                  <div className="min-w-0">
                    <div className="text-[10px] font-semibold tracking-[0.18em] text-slate-400 uppercase">
                      当前账号
                    </div>
                    <div className="truncate text-lg font-semibold text-slate-950">
                      {user.username}
                    </div>
                    <div className="mt-1 inline-flex items-center rounded-full bg-slate-100 px-2.5 py-1 text-[11px] font-semibold text-slate-600">
                      {userRoleLabel(user.role)}
                    </div>
                  </div>
                </div>
                <Button
                  variant="outline"
                  className="h-11 rounded-[18px] border-slate-200 bg-white px-4 shadow-none hover:border-slate-300 hover:bg-slate-50"
                  disabled={isLogoutPending}
                  onClick={() => void logout()}
                >
                  {isLogoutPending ? (
                    <LoaderCircle className="size-4 animate-spin" />
                  ) : (
                    <LogOut className="size-4" />
                  )}
                  {isLogoutPending ? "退出中..." : "退出登录"}
                </Button>
              </div>
            </div>
          </header>
        ) : null}

        <main className={cn("flex-1", showNavigation ? "py-5" : "py-2 sm:py-3")}>
          <Outlet />
        </main>
      </div>
    </div>
  )
}

function resolveUserMonogram(username: string) {
  const trimmed = username.trim()
  if (!trimmed) {
    return "U"
  }

  return trimmed.slice(0, 1).toUpperCase()
}
