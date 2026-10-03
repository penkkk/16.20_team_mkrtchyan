import type { PropsWithChildren } from 'react'
import { AuthBackground } from './auth-background'

export function AuthLayout({ children }: PropsWithChildren) {
  return (
    <main className="isolate relative grid min-h-svh place-items-center overflow-hidden bg-background px-5 py-10 text-foreground">
      <AuthBackground />
      {children}
    </main>
  )
}
