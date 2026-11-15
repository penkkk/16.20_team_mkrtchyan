import { HeroSearch } from "./home-hero-search";
import { ArrowDown } from 'lucide-react'
import { useEffect, useState } from 'react'

const quickTopics = ['React', 'Python', 'DevOps', 'Алгоритмы', 'Базы данных']
const headingText = 'Ваша библиотека знаний в мире IT'
const headingCharacters = Array.from(headingText)

function TypewriterHeading() {
  const [visibleCount, setVisibleCount] = useState(() =>
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
      ? headingCharacters.length
      : 0,
  )

  useEffect(() => {
    const motionPreference = window.matchMedia('(prefers-reduced-motion: reduce)')
    const showFullHeading = () => {
      if (motionPreference.matches) setVisibleCount(headingCharacters.length)
    }

    motionPreference.addEventListener('change', showFullHeading)
    return () => motionPreference.removeEventListener('change', showFullHeading)
  }, [])

  useEffect(() => {
    if (
      visibleCount >= headingCharacters.length ||
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    ) return

    const timeout = window.setTimeout(
      () => setVisibleCount(visibleCount + 1),
      visibleCount === 0 ? 250 : 125,
    )
    return () => window.clearTimeout(timeout)
  }, [visibleCount])

  return (
    <h1
      aria-label={headingText}
      className="pixel-heading relative max-w-5xl text-balance text-5xl leading-[1.35] font-medium sm:text-5xl lg:text-7xl"
    >
      <span aria-hidden="true" className="invisible block">{headingText}</span>
      <span aria-hidden="true" className="absolute inset-0 block">
        {headingCharacters.slice(0, visibleCount).join('')}
        {visibleCount < headingCharacters.length ? <span className="typewriter-caret" /> : null}
      </span>
    </h1>
  )
}

export function HomeHero() {
  return (
    <section className="relative z-10 mx-auto flex min-h-svh max-w-6xl flex-col items-center justify-center px-6 pb-24 pt-28 text-center sm:px-10">
      <TypewriterHeading />
      <p className="mt-5 max-w-md text-base leading-7 text-muted-foreground">
        Находите книги, авторов и технологии для учёбы и командной работы
      </p>
      <HeroSearch />
      <div className="mt-5 flex flex-wrap items-center justify-center gap-2">
        <span className="mr-1 text-sm text-muted-foreground">
          Популярные темы:
        </span>
        {quickTopics.map((topic) => (
          <a
            className="pixel-button border border-border bg-muted px-3 py-1.5 text-sm text-foreground transition-colors hover:bg-accent"
            href={`/catalog?${new URLSearchParams({ q: topic }).toString()}`}
            key={topic}
          >
            {topic}
          </a>
        ))}
      </div>
      <a
        className="group mt-12 inline-flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        href="#featured"
      >
        Популярные книги
        <span className="grid size-6 place-items-center rounded-full border border-border bg-muted transition-transform duration-200 group-hover:translate-y-1">
          <ArrowDown aria-hidden="true" className="size-3.5" />
        </span>
      </a>      
    </section>
    
  )
}
