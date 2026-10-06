import { createFileRoute } from '@tanstack/react-router'
import { HomePage } from '@/features/home/ui/home-page'

export const Route = createFileRoute('/')({
  component: IndexRoute,
})

function IndexRoute() {
  return <HomePage />
}
