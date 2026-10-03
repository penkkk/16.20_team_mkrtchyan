import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/profile')({
  component: ProfileRoute,
})

function ProfileRoute() {
  return <main className="min-h-svh bg-background p-8 text-foreground">Профиль ITLib</main>
}
