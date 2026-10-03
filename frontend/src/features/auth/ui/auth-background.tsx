import { cn } from '@/lib/utils'

type AuthBackgroundProps = {
  className?: string
}

export function AuthBackground({ className }: AuthBackgroundProps) {
  return (
    <div
      aria-hidden="true"
      className={cn(
        'pointer-events-none absolute inset-0 -z-10 [background:radial-gradient(125%_125%_at_50%_10%,var(--itlib-canvas)_40%,var(--itlib-ambient-violet)_100%)]',
        className,
      )}
    />
  )
}
