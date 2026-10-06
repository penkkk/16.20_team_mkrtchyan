import { HeroSearch } from "./home-hero-search";
import { ArrowDown } from 'lucide-react'

const quickTopics = ['React', 'Python', 'DevOps', 'Алгоритмы', 'Базы данных']

export function HomeHero() {
  return (
    <section className="mx-auto flex flex-col min-h-[calc(100svh-104px)] max-w-6xl items-center justify-center px-6 pb-24 pt-12 text-center sm:px-10">
      <h1 className="glass-heading max-w-5xl text-balance text-5xl leading-[0.98] font-medium sm:text-6xl lg:text-7xl">
        Ваша библиотека знаний в мире IT
      </h1>
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
            className="rounded-full border border-border bg-muted/70 px-3 py-1.5 text-sm text-foreground transition-colors hover:border-primary/60 hover:bg-primary/15"
            href={`/catalog?${new URLSearchParams({ q: topic }).toString()}`}
            key={topic}
          >
            {topic}
          </a>
        ))}
      </div>
      <a
        className="group mt-12 inline-flex items-center gap-2 rounded-full px-3 py-2 text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
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
