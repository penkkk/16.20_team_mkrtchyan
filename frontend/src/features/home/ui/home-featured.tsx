import { Card, CardContent } from "@/components/ui/card"
import { useEffect, useState } from "react"
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from "@/components/ui/carousel"
import { Button } from '@/components/ui/button'
import { ArrowRight, BookOpenText } from 'lucide-react'
import { type CarouselApi } from "@/components/ui/carousel"

const featuredBooks: Array<{
  title: string
  author: string
  description: string;
  urlimg: string;
  codeName: string;

}> = [
  {
    title: 'Чистый код',
    author: 'Роберт Мартин',
    description: 'Практическое руководство о том, как писать понятный, поддерживаемый и надёжный код.',
    urlimg: 'https://imo10.labirint.ru/books/642466/cover.jpg/484-0',
    codeName: 'Clean Code',
  },
  {
    title: 'Совершенный код',
    author: 'Стив Макконнелл',
    description: 'Подробный разбор подходов и привычек, которые помогают создавать качественные программы.',
    urlimg: 'https://ir.ozone.ru/s3/multimedia-6/6235097646.jpg',
    codeName: 'Архитектура',
  },
  {
    title: 'Грокаем алгоритмы',
    author: 'Адитья Бхаргава',
    description: 'Наглядное введение в ключевые алгоритмы и структуры данных с простыми примерами.',
    urlimg: 'https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTzPkkdsdTX0bM-Aanr46MlScQpsN1CN87y0Zfft8SOlpLX92p-nEh_7jA&s=10',
    codeName: 'Алгоритмы',
  },
  {
    title: 'Высоконагруженные приложения',
    author: 'Мартин Клеппманн',
    description: 'О проектировании масштабируемых систем, данных и надёжной архитектуры сервисов.',
    urlimg: 'https://ir.ozone.ru/s3/multimedia-g/6276358696.jpg',
    codeName: 'Архитектура'
  },
  {
    title: 'Рефакторинг',
    author: 'Мартин Фаулер',
    description: 'Каталог приёмов для безопасного улучшения кода без изменения его поведения.',
    urlimg: 'https://ir.ozone.ru/s3/multimedia-p/c1000/6506985901.jpg',
    codeName: 'Clean Code',
  },
]



export function Featured() {
  const [api, setApi] = useState<CarouselApi>()
  const [current, setCurrent] = useState(1)
  const count = featuredBooks.length

  useEffect(() => {
    if (!api) {
      return
    }

    const handleSelect = () => {
      setCurrent(api.selectedScrollSnap() + 1)
    }

    api.on("select", handleSelect)

    return () => {
      api.off("select", handleSelect)
    }
  }, [api])

  return (
    <section className="relative z-10 flex flex-col min-h-svh items-center justify-center px-6 py-16 sm:px-12" id="featured">
      <h2 className="pixel-heading max-w-5xl text-balance text-2xl leading-[1.3] font-medium lg:text-3xl">Популярные книги</h2>
      <p className="mt-3 mb-5 max-w-md text-base leading-7 text-muted-foreground text-center">Книги, которые помогут углубить знания, освоить новые технологии и писать лучший код.</p>
      <Carousel className="w-full max-w-4xl" setApi={setApi}
        opts={{
          align: "start",
          loop: true,
        }}>
        <CarouselContent>
          {featuredBooks.map((book) => (
            <CarouselItem key={book.title}>
              <div className="p-1">
                <Card className="pixel-panel group overflow-hidden py-0 transition-transform duration-200 sm:hover:-translate-y-1">
                  <CardContent className="grid overflow-hidden p-0 md:min-h-[30rem] md:grid-cols-[minmax(17rem,25rem)_minmax(0,1fr)]">
                    <div className="relative min-h-[18rem] bg-secondary sm:min-h-[22rem] md:min-h-0">
                      <img
                        alt={`Обложка книги «${book.title}»`}
                        className="size-full object-cover"
                        loading="lazy"
                        src={book.urlimg}
                      />
                      <div className="absolute bottom-4 left-4 inline-flex items-center gap-2 border-2 border-border bg-card px-3 py-1.5 text-xs font-medium text-foreground">
                        <BookOpenText aria-hidden="true" className="size-3.5 text-(--itlib-icon-color)" />
                        {book.codeName}
                      </div>
                    </div>

                    <div className="flex flex-col justify-between gap-8 p-6 sm:p-8 lg:p-10">
                      <div>
                        <p className="text-sm font-medium tracking-wide text-foreground">{book.author}</p>
                        <div className="mt-4 h-px w-12 bg-primary/70" />
                        <h2 className="mt-5 max-w-lg text-3xl font-semibold tracking-tight text-foreground lg:text-4xl">
                          {book.title}
                        </h2>
                        <p className="mt-5 max-w-xl text-base leading-7 text-muted-foreground">
                          {book.description}
                        </p>
                      </div>
                      <Button aria-label={`Подробнее о книге «${book.title}»`} className="h-11 w-fit px-5 text-base" type="button">
                        Подробнее
                        <ArrowRight aria-hidden="true" className="size-4 transition-transform duration-200 group-hover/button:translate-x-0.5" />
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </CarouselItem>
          ))}
        </CarouselContent>
        <p aria-live="polite" className="mt-5 text-center text-sm font-medium tracking-wide text-muted-foreground">
          Книга {current} из {count}
        </p>
        <CarouselPrevious className="left-2 border-border bg-card text-foreground hover:bg-muted sm:-left-12" />
        <CarouselNext className="right-2 border-border bg-card text-foreground hover:bg-muted sm:-right-12" />
      </Carousel>
    </section>
  )
}
