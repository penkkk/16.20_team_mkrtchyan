import { createFileRoute } from '@tanstack/react-router'
import { AdminPage } from '@/features/admin/ui/admin-page'

export const Route = createFileRoute('/admin')({
  component: AdminPage,
})
