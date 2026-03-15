import {
  createRootRoute,
  createRoute,
  createRouter,
  redirect,
} from "@tanstack/react-router"

import { App } from "@/App"
import { AccountsPage } from "@/features/accounts/accounts-page"
import { CardDetailPage } from "@/features/cards/card-detail-page"
import { CardsPage } from "@/features/cards/cards-page"
import { MailboxesPage } from "@/features/mailboxes/mailboxes-page"
import { UsersPage } from "@/features/users/users-page"

const rootRoute = createRootRoute({
  component: App,
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  beforeLoad: () => {
    throw redirect({ to: "/login" })
  },
})

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/login",
  component: () => null,
})

const cardsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/cards",
  component: CardsPage,
})

const cardDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/cards/$cardId",
  component: CardDetailPage,
})

const accountsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/accounts",
  component: AccountsPage,
})

const usersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/users",
  component: UsersPage,
})

const mailboxesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/mailboxes",
  component: MailboxesPage,
})

const routeTree = rootRoute.addChildren([
  indexRoute,
  loginRoute,
  cardsRoute,
  cardDetailRoute,
  accountsRoute,
  mailboxesRoute,
  usersRoute,
])

export const router = createRouter({
  routeTree,
  defaultPreload: "intent",
})

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router
  }
}
