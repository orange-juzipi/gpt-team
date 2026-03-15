import { useEffect, useState, type FormEvent } from "react"
import { LogIn } from "lucide-react"

import { Button } from "@workspace/ui/components/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@workspace/ui/components/card"
import { Input } from "@workspace/ui/components/input"
import { Label } from "@workspace/ui/components/label"

import { useMessage } from "@/components/message-context"
import { useAuth } from "@/features/auth/auth-provider"

const REMEMBERED_CREDENTIALS_KEY = "gpt-team:remembered-credentials"

export function LoginPage() {
  const { login, isLoginPending } = useAuth()
  const message = useMessage()
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [rememberPassword, setRememberPassword] = useState(false)

  useEffect(() => {
    const rememberedCredentials = window.localStorage.getItem(
      REMEMBERED_CREDENTIALS_KEY
    )
    if (!rememberedCredentials) {
      return
    }

    try {
      const parsed = JSON.parse(rememberedCredentials) as {
        username?: string
        password?: string
      }

      setUsername(parsed.username ?? "")
      setPassword(parsed.password ?? "")
      setRememberPassword(Boolean(parsed.username || parsed.password))
    } catch {
      window.localStorage.removeItem(REMEMBERED_CREDENTIALS_KEY)
    }
  }, [])

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    try {
      await login({ username, password })
      if (rememberPassword && (username.trim() || password)) {
        window.localStorage.setItem(
          REMEMBERED_CREDENTIALS_KEY,
          JSON.stringify({
            username: username.trim(),
            password,
          })
        )
      } else {
        window.localStorage.removeItem(REMEMBERED_CREDENTIALS_KEY)
      }
    } catch (error) {
      message.error(error instanceof Error ? error.message : "登录失败")
    }
  }

  return (
    <div className="flex min-h-svh items-center justify-center bg-[radial-gradient(circle_at_top_left,_rgba(190,242,100,0.32),_transparent_24%),radial-gradient(circle_at_bottom_right,_rgba(125,211,252,0.28),_transparent_26%),linear-gradient(180deg,_#fffef7_0%,_#f3f7fb_50%,_#e9f1ff_100%)] px-4 py-8">
      <div className="w-full max-w-[1040px] rounded-[36px] border border-white/70 bg-white/70 p-4 shadow-[0_32px_90px_rgba(15,23,42,0.12)] backdrop-blur sm:p-6">
        <div className="grid gap-5 lg:grid-cols-[1.15fr_0.85fr]">
          <Card className="overflow-hidden border-white/80 bg-[linear-gradient(135deg,_rgba(15,23,42,0.96)_0%,_rgba(35,44,67,0.93)_48%,_rgba(18,96,143,0.9)_100%)] text-white shadow-none">
            <CardHeader className="relative space-y-5 border-b border-white/12 pb-6">
              <div className="absolute top-6 right-6 h-28 w-28 rounded-full bg-white/8 blur-2xl" />
              <div className="absolute right-16 bottom-0 h-20 w-20 rounded-full bg-sky-300/10 blur-2xl" />
              <div className="inline-flex w-fit items-center rounded-full border border-white/15 bg-white/10 px-3 py-1 text-xs tracking-[0.18em] uppercase">
                GPT Team Console
              </div>
              <div className="relative space-y-3">
                <CardTitle className="max-w-lg text-4xl font-semibold tracking-tight sm:text-5xl">
                  欢迎回来
                </CardTitle>
                <CardDescription className="max-w-xl text-base leading-8 text-slate-200">
                  输入账号与密码，继续进入工作台。
                </CardDescription>
              </div>
            </CardHeader>
            <CardContent className="grid gap-4 pt-6 sm:grid-cols-[1.1fr_0.9fr]">
              <div className="rounded-[28px] border border-white/12 bg-white/6 p-6">
                <div className="text-xs tracking-[0.2em] text-slate-300 uppercase">
                  Access
                </div>
                <div className="mt-4 text-2xl font-semibold text-white">
                  Sign in to continue
                </div>
                <div className="mt-3 text-sm leading-7 text-slate-200">
                  保持专注、安静、直接的登录入口。
                </div>
              </div>

              <div className="rounded-[28px] border border-white/12 bg-[linear-gradient(180deg,_rgba(255,255,255,0.08),_rgba(255,255,255,0.03))] p-6">
                <div className="grid grid-cols-3 gap-3">
                  {Array.from({ length: 9 }).map((_, index) => (
                    <div
                      key={index}
                      className="aspect-square rounded-2xl border border-white/10 bg-white/6"
                    />
                  ))}
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="border-white/80 bg-white/92 shadow-[0_20px_60px_rgba(148,163,184,0.16)]">
            <CardHeader>
              <CardTitle>登录系统</CardTitle>
              <CardDescription>使用用户名和密码进入后台。</CardDescription>
            </CardHeader>
            <CardContent>
              <form className="space-y-4" onSubmit={(event) => void handleSubmit(event)}>
                <Field label="用户名">
                  <Input
                    aria-label="用户名"
                    autoComplete="username"
                    value={username}
                    onChange={(event) => setUsername(event.target.value)}
                  />
                </Field>
                <Field label="密码">
                  <Input
                    aria-label="密码"
                    type="password"
                    autoComplete="current-password"
                    value={password}
                    onChange={(event) => setPassword(event.target.value)}
                  />
                </Field>

                <label className="flex items-center gap-2 text-sm text-slate-600">
                  <input
                    aria-label="记住密码"
                    type="checkbox"
                    checked={rememberPassword}
                    onChange={(event) => {
                      const checked = event.target.checked
                      setRememberPassword(checked)
                      if (!checked) {
                        window.localStorage.removeItem(REMEMBERED_CREDENTIALS_KEY)
                      }
                    }}
                  />
                  记住密码
                </label>

                <Button className="w-full" type="submit" disabled={isLoginPending}>
                  <LogIn className="size-4" />
                  {isLoginPending ? "登录中..." : "登录"}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

function Field({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      {children}
    </div>
  )
}
