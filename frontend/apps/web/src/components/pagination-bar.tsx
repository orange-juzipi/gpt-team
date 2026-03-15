import { ChevronLeft, ChevronRight } from "lucide-react"

import { Button } from "@workspace/ui/components/button"
import { Select } from "@workspace/ui/components/select"

const PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const

export function PaginationBar({
  page,
  totalPages,
  totalItems,
  currentCount,
  pageSize,
  onPageChange,
  onPageSizeChange,
}: {
  page: number
  totalPages: number
  totalItems: number
  currentCount: number
  pageSize: number
  onPageChange: (page: number) => void
  onPageSizeChange: (pageSize: number) => void
}) {
  if (totalItems <= 0) {
    return null
  }

  return (
    <div className="flex flex-col gap-3 rounded-[24px] border border-white/70 bg-white/75 px-4 py-3 shadow-sm sm:flex-row sm:items-center sm:justify-between">
      <div className="flex flex-wrap items-center gap-3">
        <div className="text-sm text-muted-foreground">
          第 {page} / {totalPages} 页，当前显示 {currentCount} 条，共 {totalItems} 条
        </div>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <span>每页</span>
          <Select
            aria-label="每页数量"
            className="h-9 w-[92px] rounded-xl px-3 py-0 shadow-none"
            value={String(pageSize)}
            onChange={(event) => onPageSizeChange(Number(event.target.value))}
          >
            {PAGE_SIZE_OPTIONS.map((option) => (
              <option key={option} value={option}>
                {option} 条
              </option>
            ))}
          </Select>
        </div>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <Button
          size="sm"
          variant="outline"
          disabled={page <= 1}
          onClick={() => onPageChange(Math.max(1, page - 1))}
        >
          <ChevronLeft className="size-4" />
          上一页
        </Button>
        <div className="flex items-center gap-1">
          {buildPageNumbers(page, totalPages).map((value, index) =>
            value === "ellipsis" ? (
              <span
                key={`ellipsis-${index}`}
                className="px-2 text-sm text-muted-foreground"
              >
                ...
              </span>
            ) : (
              <Button
                key={value}
                size="sm"
                variant={value === page ? "default" : "outline"}
                onClick={() => onPageChange(value)}
              >
                {value}
              </Button>
            )
          )}
        </div>
        <Button
          size="sm"
          variant="outline"
          disabled={page >= totalPages}
          onClick={() => onPageChange(Math.min(totalPages, page + 1))}
        >
          下一页
          <ChevronRight className="size-4" />
        </Button>
      </div>
    </div>
  )
}

function buildPageNumbers(page: number, totalPages: number) {
  if (totalPages <= 5) {
    return Array.from({ length: totalPages }, (_, index) => index + 1)
  }

  if (page <= 3) {
    return [1, 2, 3, 4, "ellipsis", totalPages] as const
  }

  if (page >= totalPages - 2) {
    return [1, "ellipsis", totalPages - 3, totalPages - 2, totalPages - 1, totalPages] as const
  }

  return [1, "ellipsis", page - 1, page, page + 1, "ellipsis", totalPages] as const
}
