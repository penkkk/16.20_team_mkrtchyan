import { HomeBackground } from './home-background'
import { HomeHeader } from './home-header'
import { HomeHero } from './home-hero'
import { CarouselDemo } from './home-featured'

export function HomePage() {
  return (
    <main className="isolate relative min-h-svh overflow-hidden bg-background text-foreground">
      <HomeBackground />
      <HomeHeader />
      <HomeHero />
      <CarouselDemo />
    </main>
  )
}
