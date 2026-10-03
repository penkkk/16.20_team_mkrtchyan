import { createFileRoute } from '@tanstack/react-router'
import { LoginPage } from '@/features/auth/ui/login-page'

export const Route = createFileRoute('/login')({
  component: LoginRoute,
})

function LoginRoute() {
  return <LoginPage />
}
