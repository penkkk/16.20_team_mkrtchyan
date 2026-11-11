import { useEffect, useState } from 'react'
import { Moon, Sun } from 'lucide-react'
import { Button } from '@/components/ui/button'

const storageKey = 'itlib-theme'

export function ThemeToggle({ className = '' }: { className?: string }) {
  const [isDark, setIsDark] = useState(() => document.documentElement.classList.contains('dark'))

  useEffect(() => {
    const sync = () => setIsDark(document.documentElement.classList.contains('dark'))
    const onStorage = (event: StorageEvent) => {
      if (event.key !== storageKey) return
      document.documentElement.classList.toggle('dark', event.newValue !== 'light')
      sync()
    }
    window.addEventListener('storage', onStorage)
    return () => window.removeEventListener('storage', onStorage)
  }, [])

  function toggleTheme() {
    const nextIsDark = !isDark
    document.documentElement.classList.toggle('dark', nextIsDark)
    setIsDark(nextIsDark)
    try {
      localStorage.setItem(storageKey, nextIsDark ? 'dark' : 'light')
    } catch {
      //
    }
  }

  return (
    <Button
      aria-label={isDark ? 'Включить светлую тему' : 'Включить тёмную тему'}
      className={`pixel-button text-foreground hover:bg-accent ${className}`}
      size="icon"
      title={isDark ? 'Светлая тема' : 'Тёмная тема'}
      type="button"
      variant="ghost"
      onClick={toggleTheme}
    >
      {isDark ? <Sun aria-hidden="true" /> : <Moon aria-hidden="true" />}
    </Button>
  )
}
