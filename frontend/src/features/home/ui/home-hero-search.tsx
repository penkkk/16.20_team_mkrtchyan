import { Search } from 'lucide-react'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from "@/components/ui/input-group"


export function HeroSearch() {
    return (
    <form
      action="/catalog"
      className="mx-auto mt-8 flex w-full max-w-2xl flex-col gap-3 sm:flex-row justify-center"
      method="get"
    >  
    <InputGroup className="h-12 rounded-3xl glass-header outline-0" >
      <InputGroupInput placeholder="Искать книгу..." name='q' className='ml-1'/>
      <InputGroupAddon>
        <Search />
      </InputGroupAddon>
      <InputGroupAddon align="inline-end">
        <kbd className="text-xs text-muted-foreground mr-2">Enter</kbd>
      </InputGroupAddon>
    </InputGroup>
    </form>
    )
}
