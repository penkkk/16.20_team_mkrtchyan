import { Link } from '@tanstack/react-router'
import { BookOpenText, Menu, Moon, Bell } from 'lucide-react'
import { Button } from '@/components/ui/button'
import ButtonWithIconDemo from '@/components/ui/button-with-icon'


const navigationItems = [
  { label: 'Каталог', href: '#catalog' },
  { label: 'Подборки', href: '#collections' },
  { label: 'О библиотеке', href: '#about' },
]

export function HomeHeader() {
  return (
    <header className="mx-auto w-full max-w-7xl px-4 pt-4 sm:px-6 sm:pt-6">
      <div className="glass-header flex min-h-14 items-center justify-between gap-3 rounded-3xl border border-border bg-[rgb(5_6_15_/_58%)] px-3 shadow-[inset_0_1px_0_rgb(216_236_248_/_10%),0_16px_40px_rgb(0_0_0_/_24%)] backdrop-blur-xl sm:min-h-16 sm:px-5">
        <Link
          className="flex items-center gap-3 rounded-lg text-sm font-semibold text-foreground outline-none transition-opacity hover:opacity-80 focus-visible:ring-2 focus-visible:ring-ring"
          to="/"
        >
          <span className="grid size-8 place-items-center rounded-xl border border-border bg-muted">
            <BookOpenText className="size-4 text-(--itlib-ice)" aria-hidden="true" />
          </span>
          <span className="text-base tracking-tight">ITLib</span>
        </Link>

        <nav aria-label="Основная навигация" className="hidden items-center gap-1 md:flex">
          {navigationItems.map(({ label, href }) => (
            <a
              className="rounded-lg px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              href={href}
              key={href}
            >
              {label}
            </a>
          ))}
        </nav>

        <div className="flex items-center gap-1.5">
          <Button
            aria-label="Уведомления"
            className="hidden text-muted-foreground hover:text-foreground sm:inline-flex"
            size="icon"
            type="button"
            variant="ghost"
          >
            <Bell aria-hidden="true" />
          </Button>
          <Button
            aria-label="Переключить тему"
            className="hidden text-muted-foreground hover:text-foreground sm:inline-flex"
            size="icon"
            type="button"
            variant="ghost"
          >
            <Moon aria-hidden="true" />
          </Button>
          <Link
            to="/login"
          >
            <ButtonWithIconDemo></ButtonWithIconDemo>
          </Link>
          <Button
            aria-label="Открыть меню"
            className="text-muted-foreground hover:text-foreground md:hidden"
            size="icon"
            type="button"
            variant="ghost"
          >
            <Menu aria-hidden="true" />
          </Button>
        </div>
      </div>
    </header>
  )
}
