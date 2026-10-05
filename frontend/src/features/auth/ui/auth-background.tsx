import { cn } from '@/lib/utils'

type AuthBackgroundProps = {
  className?: string
}

export function AuthBackground({ className }: AuthBackgroundProps) {
  return (
    <div
      aria-hidden="true"
      className={cn(
        'auth-background pointer-events-none absolute inset-0 -z-10 [background:radial-gradient(85%_115%_at_72%_-12%,rgb(182_217_252_/_30%)_0%,rgb(2_125_234_/_18%)_43%,rgb(15_76_151_/_12%)_68%,transparent_92%),radial-gradient(70%_60%_at_72%_65%,rgb(2_125_234_/_13%)_0%,transparent_78%),linear-gradient(180deg,#101a33_0%,#090f20_58%,var(--itlib-canvas)_100%)]',
        className,
      )}
    />
  )
}
