import { useEffect, useMemo, useState } from "react"

export const DEFAULT_PAGE_SIZE = 10

export function usePagination<T>({
  items,
  resetKeys = [],
  initialPageSize = DEFAULT_PAGE_SIZE,
}: {
  items: T[]
  resetKeys?: readonly unknown[]
  initialPageSize?: number
}) {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(initialPageSize)

  const totalPages = useMemo(
    () => Math.max(1, Math.ceil(items.length / pageSize)),
    [items.length, pageSize]
  )

  const paginatedItems = useMemo(() => {
    const startIndex = (page - 1) * pageSize
    return items.slice(startIndex, startIndex + pageSize)
  }, [items, page, pageSize])

  useEffect(() => {
    setPage(1)
  }, [pageSize, ...resetKeys])

  useEffect(() => {
    if (page > totalPages) {
      setPage(totalPages)
    }
  }, [page, totalPages])

  return {
    page,
    setPage,
    pageSize,
    setPageSize,
    totalPages,
    paginatedItems,
  }
}
