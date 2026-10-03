import { createFileRoute } from '@tanstack/react-router'
import { RegisterPage } from '@/features/auth/ui/register-page'

export const Route = createFileRoute('/register')({
  component: RegisterRoute,
})

function RegisterRoute() {
  return <RegisterPage />
}
