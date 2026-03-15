import { createContext, useContext } from "react"

export type MessageTone = "success" | "error" | "info"

export type MessageAPI = {
  show: (tone: MessageTone, text: string) => void
  success: (text: string) => void
  error: (text: string) => void
  info: (text: string) => void
}

export const MessageContext = createContext<MessageAPI | null>(null)

export function useMessage() {
  const context = useContext(MessageContext)
  if (!context) {
    throw new Error("useMessage must be used within <MessageProvider>")
  }

  return context
}
