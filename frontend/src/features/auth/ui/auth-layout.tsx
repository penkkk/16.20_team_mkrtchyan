import type { PropsWithChildren } from 'react'
import { AuthBackground } from './auth-background'
import { ThemeToggle } from '@/components/theme-toggle'

export function AuthLayout({ children }: PropsWithChildren) {
  return (
    <main className="isolate relative grid min-h-svh place-items-center overflow-hidden bg-background px-5 py-10 text-foreground">
      <AuthBackground />
      <ThemeToggle className="absolute right-5 top-5 z-20" />
      {children}
    </main>
  )
}
