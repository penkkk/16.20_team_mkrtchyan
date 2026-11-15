import { createFileRoute } from '@tanstack/react-router'
import { ThemeToggle } from '@/components/theme-toggle'

export const Route = createFileRoute('/profile')({
  component: ProfileRoute,
})

function ProfileRoute() {
  return (
    <main className="flex min-h-svh items-start justify-between bg-background p-8 text-foreground">
      <h1 className="font-heading text-2xl">Профиль ITLib</h1>
      <ThemeToggle />
    </main>
  )
}
