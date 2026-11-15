import { cn } from '@/lib/utils'

type AuthBackgroundProps = {
  className?: string
}

export function AuthBackground({ className }: AuthBackgroundProps) {
  return (
    <div
      aria-hidden="true"
      className={cn(
        'auth-background pointer-events-none absolute inset-0 -z-10',
        className,
      )}
    />
  )
}
