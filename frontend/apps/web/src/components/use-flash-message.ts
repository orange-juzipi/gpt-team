import { useEffect, useRef } from "react"

import { useMessage, type MessageTone } from "@/components/message-context"

export function useFlashMessage(
  text: string | null | undefined,
  tone: MessageTone = "error"
) {
  const message = useMessage()
  const lastTextRef = useRef<string | null>(null)

  useEffect(() => {
    if (!text) {
      lastTextRef.current = null
      return
    }

    if (lastTextRef.current === text) {
      return
    }

    lastTextRef.current = text
    message.show(tone, text)
  }, [message, text, tone])
}
