import {
  createContext,
  useContext,
  type ReactNode,
} from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { ApiRequestError, api } from "@/lib/api"
import type { AuthUser, LoginPayload } from "@/lib/types"

type AuthContextValue = {
  user: AuthUser | null
  isLoading: boolean
  error: Error | null
  retry: () => Promise<unknown>
  login: (payload: LoginPayload) => Promise<AuthUser>
  logout: () => Promise<void>
  isLoginPending: boolean
  isLogoutPending: boolean
  isRetryPending: boolean
}

const AUTH_QUERY_KEY = ["auth", "me"] as const

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient()

  const authQuery = useQuery({
    queryKey: AUTH_QUERY_KEY,
    queryFn: async () => {
      try {
        return await api.getCurrentUser()
      } catch (error) {
        if (error instanceof ApiRequestError && error.status === 401) {
          return null
        }

        throw error
      }
    },
    retry: false,
  })

  const loginMutation = useMutation({
    mutationFn: (payload: LoginPayload) => api.login(payload),
    onSuccess: (user) => {
      queryClient.setQueryData(AUTH_QUERY_KEY, user)
    },
  })

  const logoutMutation = useMutation({
    mutationFn: api.logout,
    onSuccess: () => {
      queryClient.removeQueries()
      queryClient.setQueryData(AUTH_QUERY_KEY, null)
    },
  })

  return (
    <AuthContext.Provider
      value={{
        user: authQuery.data ?? null,
        isLoading: authQuery.isLoading,
        error: authQuery.error,
        retry: authQuery.refetch,
        login: loginMutation.mutateAsync,
        logout: logoutMutation.mutateAsync,
        isLoginPending: loginMutation.isPending,
        isLogoutPending: logoutMutation.isPending,
        isRetryPending: authQuery.isFetching,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error("useAuth must be used within <AuthProvider>")
  }

  return context
}
