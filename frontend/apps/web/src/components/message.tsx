import {
  type ReactNode,
  useMemo,
  useEffect,
  useRef,
  useState,
} from "react"

import { cn } from "@workspace/ui/lib/utils"
import { MessageContext, type MessageAPI, type MessageTone } from "@/components/message-context"

type MessageItem = {
  id: number
  tone: MessageTone
  text: string
  state: "enter" | "visible" | "exit"
}
const MESSAGE_DURATION_MS = 2200
const MESSAGE_EXIT_DURATION_MS = 220

export function MessageProvider({ children }: { children: ReactNode }) {
  const [messages, setMessages] = useState<MessageItem[]>([])
  const nextIdRef = useRef(1)
  const timeoutsRef = useRef<Map<number, number[]>>(new Map())
  const rafsRef = useRef<Map<number, number>>(new Map())

  const api = useMemo<MessageAPI>(() => {
    const push = (tone: MessageTone, text: string) => {
      const id = nextIdRef.current++
      setMessages((current) => [...current, { id, tone, text, state: "enter" }])

      const rafId = window.requestAnimationFrame(() => {
        setMessages((current) =>
          current.map((message) =>
            message.id === id ? { ...message, state: "visible" } : message
          )
        )
        rafsRef.current.delete(id)
      })
      rafsRef.current.set(id, rafId)

      const exitTimer = window.setTimeout(() => {
        setMessages((current) =>
          current.map((message) =>
            message.id === id ? { ...message, state: "exit" } : message
          )
        )

        const removeTimer = window.setTimeout(() => {
          setMessages((current) => current.filter((message) => message.id !== id))
          timeoutsRef.current.delete(id)
        }, MESSAGE_EXIT_DURATION_MS)

        const timers = timeoutsRef.current.get(id) ?? []
        timeoutsRef.current.set(id, [...timers, removeTimer])
      }, MESSAGE_DURATION_MS)

      timeoutsRef.current.set(id, [exitTimer])
    }

    return {
      show: push,
      success: (text: string) => push("success", text),
      error: (text: string) => push("error", text),
      info: (text: string) => push("info", text),
    }
  }, [])

  useEffect(() => {
    return () => {
      rafsRef.current.forEach((rafId) => window.cancelAnimationFrame(rafId))
      rafsRef.current.clear()
      timeoutsRef.current.forEach((timers) => {
        timers.forEach((timer) => window.clearTimeout(timer))
      })
      timeoutsRef.current.clear()
    }
  }, [])

  return (
    <MessageContext.Provider value={api}>
      {children}
      <div className="pointer-events-none fixed inset-x-0 top-4 z-[80] flex flex-col items-center gap-2 px-4">
        {messages.map((message) => (
          <div
            key={message.id}
            className={cn(
              "min-w-[220px] max-w-[420px] rounded-2xl border px-4 py-3 text-sm shadow-[0_16px_36px_rgba(15,23,42,0.16)] backdrop-blur transition-all duration-200 ease-out",
              message.state === "enter" && "-translate-y-2 scale-95 opacity-0",
              message.state === "visible" && "translate-y-0 scale-100 opacity-100",
              message.state === "exit" && "-translate-y-2 scale-95 opacity-0",
              message.tone === "success" &&
                "border-emerald-200 bg-emerald-50/95 text-emerald-800",
              message.tone === "error" &&
                "border-rose-200 bg-rose-50/95 text-rose-800",
              message.tone === "info" &&
                "border-sky-200 bg-sky-50/95 text-sky-800"
            )}
          >
            {message.text}
          </div>
        ))}
      </div>
    </MessageContext.Provider>
  )
}
