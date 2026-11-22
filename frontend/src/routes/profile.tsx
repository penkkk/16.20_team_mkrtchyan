import { Link, Navigate, createFileRoute, useNavigate } from '@tanstack/react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { BookOpenText, LogOut } from 'lucide-react'
import { ThemeToggle } from '@/components/theme-toggle'
import { Button } from '@/components/ui/button'
import { currentUserQueryKey, getCurrentUser, logout } from '@/features/auth/api/auth-api'

export const Route = createFileRoute('/profile')({
  component: ProfileRoute,
})

function ProfileRoute() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const userQuery = useQuery({
    queryKey: currentUserQueryKey,
    queryFn: getCurrentUser,
    retry: false,
    staleTime: 0,
  })
  const logoutMutation = useMutation({
    mutationFn: logout,
    onSuccess: () => {
      queryClient.setQueryData(currentUserQueryKey, null)
      void navigate({ to: '/login' })
    },
  })

  if (userQuery.isPending) {
    return <ProfileShell><p className="text-muted-foreground">Восстанавливаем сессию...</p></ProfileShell>
  }

  if (userQuery.isError) {
    return (
      <ProfileShell>
        <div className="pixel-panel max-w-lg rounded-2xl border p-6">
          <h1 className="font-heading text-2xl">Не удалось загрузить профиль</h1>
          <p className="mt-3 text-sm text-muted-foreground">
            {userQuery.error instanceof Error ? userQuery.error.message : 'Попробуйте еще раз.'}
          </p>
          <Button className="mt-5" onClick={() => void userQuery.refetch()}>Повторить</Button>
        </div>
      </ProfileShell>
    )
  }

  if (!userQuery.data) return <Navigate to="/login" />

  const user = userQuery.data
  const fullName = [user.name, user.surname].filter(Boolean).join(' ') || user.username

  return (
    <ProfileShell>
      <section className="pixel-panel w-full max-w-xl rounded-2xl border p-6 sm:p-8">
        <div className="flex items-start gap-4">
          <div className="grid size-12 shrink-0 place-items-center rounded-xl border border-border bg-muted">
            <BookOpenText className="size-6 text-(--itlib-icon-color)" aria-hidden="true" />
          </div>
          <div>
            <p className="text-sm font-medium tracking-[0.14em] text-muted-foreground uppercase">ITLib</p>
            <h1 className="mt-1 font-heading text-2xl text-foreground">Профиль</h1>
          </div>
        </div>

        <dl className="mt-8 grid gap-5 text-sm">
          <div>
            <dt className="text-muted-foreground">Имя</dt>
            <dd className="mt-1 font-medium text-foreground">{fullName}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Логин</dt>
            <dd className="mt-1 font-medium text-foreground">{user.username}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Email</dt>
            <dd className="mt-1 font-medium text-foreground">{user.email}</dd>
          </div>
        </dl>

        {logoutMutation.isError ? (
          <p className="mt-6 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {logoutMutation.error instanceof Error ? logoutMutation.error.message : 'Не удалось выйти из аккаунта.'}
          </p>
        ) : null}

        <Button
          className="mt-8"
          disabled={logoutMutation.isPending}
          variant="outline"
          onClick={() => logoutMutation.mutate()}
        >
          <LogOut className="size-4" aria-hidden="true" />
          {logoutMutation.isPending ? 'Выходим...' : 'Выйти'}
        </Button>
      </section>
    </ProfileShell>
  )
}

function ProfileShell({ children }: { children: React.ReactNode }) {
  return (
    <main className="min-h-svh bg-background px-5 py-8 text-foreground sm:px-8">
      <header className="mx-auto flex w-full max-w-5xl items-center justify-between">
        <Link className="text-sm font-medium text-muted-foreground hover:text-foreground" to="/">← На главную</Link>
        <ThemeToggle />
      </header>
      <div className="mx-auto mt-16 flex w-full max-w-5xl justify-center">{children}</div>
    </main>
  )
}
