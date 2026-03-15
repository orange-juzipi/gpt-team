import { useEffect, useRef, useState, type ReactNode } from "react"
import { createPortal } from "react-dom"

import { Button } from "@workspace/ui/components/button"
import { cn } from "@workspace/ui/lib/utils"

export function HoverHintButton({
  hint,
  className,
  children,
  ...props
}: React.ComponentProps<typeof Button> & {
  hint: string
  children: ReactNode
}) {
  const wrapperRef = useRef<HTMLDivElement | null>(null)
  const [open, setOpen] = useState(false)
  const [position, setPosition] = useState<{ top: number; left: number } | null>(null)

  useEffect(() => {
    if (!open) {
      return
    }

    const updatePosition = () => {
      const rect = wrapperRef.current?.getBoundingClientRect()
      if (!rect) {
        return
      }

      setPosition({
        top: Math.max(12, rect.top - 10),
        left: rect.left + rect.width / 2,
      })
    }

    updatePosition()

    window.addEventListener("scroll", updatePosition, true)
    window.addEventListener("resize", updatePosition)

    return () => {
      window.removeEventListener("scroll", updatePosition, true)
      window.removeEventListener("resize", updatePosition)
    }
  }, [open])

  return (
    <div
      ref={wrapperRef}
      className="inline-flex"
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
      onFocus={() => setOpen(true)}
      onBlur={() => setOpen(false)}
    >
      <Button
        aria-label={hint}
        className={className}
        {...props}
      >
        {children}
      </Button>
      {open && position
        ? createPortal(
            <span
              className={cn(
                "pointer-events-none fixed z-[120] whitespace-nowrap rounded-md bg-slate-900 px-2 py-1 text-xs font-medium text-white shadow-lg"
              )}
              style={{
                top: position.top,
                left: position.left,
                transform: "translate(-50%, -100%)",
              }}
            >
              {hint}
            </span>,
            document.body
          )
        : null}
    </div>
  )
}
