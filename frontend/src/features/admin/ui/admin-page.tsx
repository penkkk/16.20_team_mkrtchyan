import { useState, type FormEvent } from 'react'
import { Link } from '@tanstack/react-router'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { BookOpenText, Download, FilePlus2, RefreshCw, Search, Trash2, UploadCloud, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ThemeToggle } from '@/components/theme-toggle'
import type { BookAccess, BookLevel } from '../model/book-store'
import { deleteBook, downloadBook, listBooks, uploadBook, type BookUpload, type CatalogBook } from '../api/book-api'

type View = 'books' | 'create'
const pageSize = 20

export function AdminPage() {
  const [view, setView] = useState<View>('books')
  const [offset, setOffset] = useState(0)
  const [searchInput, setSearchInput] = useState('')
  const [topicInput, setTopicInput] = useState('')
  const [filters, setFilters] = useState({ search: '', topic: '' })
  const [notice, setNotice] = useState('')
  const [saving, setSaving] = useState(false)
  const [actionError, setActionError] = useState('')
  const [busyBook, setBusyBook] = useState<string | null>(null)
  const queryClient = useQueryClient()
  const booksQuery = useQuery({
    queryKey: ['admin-books', offset, filters.search, filters.topic],
    queryFn: () => listBooks(offset, filters.search, filters.topic),
    enabled: view === 'books',
    retry: false,
  })

  function applyFilters(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setOffset(0)
    setFilters({ search: searchInput.trim(), topic: topicInput.trim() })
  }

  async function handleCreate(book: BookUpload) {
    setSaving(true)
    try {
      await uploadBook(book)
      setOffset(0)
      setSearchInput('')
      setTopicInput('')
      setFilters({ search: '', topic: '' })
      void queryClient.invalidateQueries({ queryKey: ['admin-books'] })
      setView('books')
      setNotice(`Книга «${book.title}» загружена на сервер.`)
    } finally {
      setSaving(false)
    }
  }

  async function handleDownload(book: CatalogBook) {
    setActionError('')
    setBusyBook(book.id)
    try { await downloadBook(book.id) } catch (error) {
      setActionError(error instanceof Error ? error.message : 'Не удалось скачать книгу.')
    } finally { setBusyBook(null) }
  }

  async function handleDelete(book: CatalogBook) {
    if (!window.confirm(`Удалить книгу «${book.title}», её файл и обложку? Это действие нельзя отменить.`)) return
    setActionError('')
    setBusyBook(book.id)
    try {
      await deleteBook(book.id)
      if (offset > 0 && booksQuery.data?.books.length === 1) setOffset((value) => Math.max(0, value - pageSize))
      void queryClient.invalidateQueries({ queryKey: ['admin-books'] })
      setNotice(`Книга «${book.title}» удалена.`)
    } catch (error) {
      setActionError(error instanceof Error ? error.message : 'Не удалось удалить книгу.')
    } finally { setBusyBook(null) }
  }

  return (
    <div className="relative isolate min-h-svh bg-background text-foreground">
      <div aria-hidden="true" className="home-background pointer-events-none absolute inset-0 -z-10" />
      <header className="border-b-2 border-border bg-card/95">
        <div className="mx-auto flex min-h-18 max-w-7xl items-center justify-between gap-4 px-5 sm:px-8">
          <div className="flex items-center gap-3">
            <Link aria-label="На главную ITLib" className="pixel-frame grid size-10 place-items-center text-foreground" to="/"><BookOpenText className="size-5" aria-hidden="true" /></Link>
            <div><p className="text-sm font-semibold tracking-wide">ITLib</p><p className="text-sm text-muted-foreground">Управление библиотекой</p></div>
          </div>
          <ThemeToggle />
        </div>
      </header>
      <main className="mx-auto max-w-7xl px-5 pb-20 pt-8 sm:px-8 sm:pt-12">
        <div className="flex flex-col justify-between gap-5 sm:flex-row sm:items-end">
          <div><p className="mb-3 text-sm font-semibold uppercase tracking-[0.16em] text-muted-foreground">Панель управления</p><h1 className="pixel-heading text-3xl font-semibold sm:text-4xl">Каталог книг</h1><p className="mt-4 max-w-2xl text-base leading-7 text-muted-foreground">Загружайте книги и просматривайте каталог библиотеки.</p></div>
          {view === 'books' && <Button className="h-11 gap-2 px-5 text-base" onClick={() => { setNotice(''); setView('create') }}><FilePlus2 className="size-5" aria-hidden="true" /> Добавить книгу</Button>}
        </div>
        {notice && <div role="status" className="mt-5 flex items-start justify-between gap-3 border-2 border-border bg-card px-4 py-3 text-sm"><span>{notice}</span><button aria-label="Скрыть сообщение" className="text-muted-foreground hover:text-foreground" onClick={() => setNotice('')} type="button"><X className="size-4" /></button></div>}
        <div className="mt-9 flex gap-2 border-b-2 border-border" aria-label="Разделы админки">
          <button disabled={saving} aria-current={view === 'books' ? 'page' : undefined} className={`px-4 py-3 text-sm font-semibold ${view === 'books' ? 'border-b-4 border-primary text-foreground' : 'text-muted-foreground hover:text-foreground'}`} onClick={() => setView('books')} type="button">Все книги</button>
          <button disabled={saving} aria-current={view === 'create' ? 'page' : undefined} className={`px-4 py-3 text-sm font-semibold ${view === 'create' ? 'border-b-4 border-primary text-foreground' : 'text-muted-foreground hover:text-foreground'}`} onClick={() => setView('create')} type="button">Добавить книгу</button>
        </div>
        {view === 'books' ? <section aria-label="Список книг" className="pixel-panel mt-6 p-5 sm:p-7">
          <div className="flex flex-wrap items-center justify-between gap-4"><div><h2 className="text-xl font-semibold">Все книги</h2><p className="mt-1 text-sm text-muted-foreground">Страница {Math.floor(offset / pageSize) + 1}</p></div><Button disabled={booksQuery.isFetching} onClick={() => void booksQuery.refetch()} type="button" variant="outline"><RefreshCw className="size-4" />Обновить</Button></div>
          <form className="mt-6 flex flex-wrap items-end gap-3" onSubmit={applyFilters}>
            <label className="grid gap-2 text-sm font-semibold">Поиск по названию, автору или теме<Input className="h-11 border-2 border-border bg-background sm:w-70" onChange={(event) => setSearchInput(event.target.value)} value={searchInput} /></label>
            <label className="grid gap-2 text-sm font-semibold">Тема<Input className="h-11 border-2 border-border bg-background sm:w-52" onChange={(event) => setTopicInput(event.target.value)} placeholder="Точное название" value={topicInput} /></label>
            <Button className="h-11" type="submit"><Search className="size-4" /> Найти</Button>
          </form>
          {actionError && <p role="alert" className="mt-5 border-2 border-destructive bg-destructive/10 px-4 py-3 text-sm text-destructive">{actionError}</p>}
          {booksQuery.isPending ? <p className="py-16 text-center text-muted-foreground">Загружаем книги...</p> : booksQuery.isError ? <div className="py-16 text-center" role="alert"><p>{booksQuery.error instanceof Error ? booksQuery.error.message : 'Не удалось загрузить каталог.'}</p><Button className="mt-4" onClick={() => void booksQuery.refetch()}>Повторить</Button></div> : booksQuery.data.books.length === 0 ? <div className="py-16 text-center"><BookOpenText className="mx-auto size-10 text-muted-foreground" aria-hidden="true" /><h3 className="mt-4 text-lg font-semibold">Книги не найдены</h3><p className="mt-2 text-sm text-muted-foreground">Измените параметры поиска или добавьте книгу.</p></div> : <div className="mt-6 overflow-x-auto"><table className="w-full min-w-150 border-collapse text-left text-sm"><thead><tr className="border-y-2 border-border bg-muted text-muted-foreground"><th className="px-4 py-3 font-semibold">Книга</th><th className="px-4 py-3 font-semibold">Тема / уровень</th><th className="px-4 py-3 font-semibold">Доступ</th><th className="px-4 py-3 text-right font-semibold">Действия</th></tr></thead><tbody>{booksQuery.data.books.map((book) => <BookRow key={book.id} book={book} disabled={busyBook !== null} onDownload={() => void handleDownload(book)} onDelete={() => void handleDelete(book)} />)}</tbody></table></div>}
          <div className="mt-6 flex items-center justify-between gap-3"><Button disabled={offset === 0 || booksQuery.isFetching} onClick={() => setOffset((value) => Math.max(0, value - pageSize))} type="button" variant="outline">Назад</Button><Button disabled={!booksQuery.data?.has_more || booksQuery.isFetching} onClick={() => setOffset((value) => value + pageSize)} type="button" variant="outline">Далее</Button></div>
        </section> : <BookForm onCancel={() => setView('books')} onCreate={handleCreate} saving={saving} />}
      </main>
    </div>
  )
}

const levelLabels: Record<CatalogBook['level'], string> = { beginner: 'Начальный', intermediate: 'Средний', advanced: 'Продвинутый' }
const accessLabels: Record<CatalogBook['access_level'], string> = { free: 'Бесплатно', subscription: 'По подписке' }

function BookRow({ book, disabled, onDownload, onDelete }: { book: CatalogBook; disabled: boolean; onDownload: () => void; onDelete: () => void }) {
  return <tr className="border-b border-border/70 align-middle hover:bg-muted/45"><td className="px-4 py-4"><div className="flex min-w-55 items-center gap-3"><span className="grid size-12 shrink-0 place-items-center border-2 border-border bg-secondary"><BookOpenText className="size-5" aria-hidden="true" /></span><div><p className="font-semibold text-foreground">{book.title}</p><p className="mt-1 text-muted-foreground">{book.author}</p></div></div></td><td className="px-4 py-4"><p>{book.topic}</p><p className="mt-1 text-muted-foreground">{levelLabels[book.level]}</p></td><td className="px-4 py-4"><span className="inline-block border border-border bg-secondary px-2 py-1 text-xs font-medium">{accessLabels[book.access_level]}</span></td><td className="px-4 py-4 text-right"><div className="flex justify-end gap-2"><Button aria-label={`Скачать книгу «${book.title}»`} disabled={disabled} onClick={onDownload} size="icon" title="Скачать книгу" type="button" variant="outline"><Download className="size-4" /></Button><Button aria-label={`Удалить книгу «${book.title}»`} disabled={disabled} onClick={onDelete} size="icon" title="Удалить книгу" type="button" variant="destructive"><Trash2 className="size-4" /></Button></div></td></tr>
}

function BookForm({ onCancel, onCreate, saving }: { onCancel: () => void; onCreate: (book: BookUpload) => Promise<void>; saving: boolean }) {
  const [title, setTitle] = useState('')
  const [author, setAuthor] = useState('')
  const [technology, setTechnology] = useState('')
  const [level, setLevel] = useState<BookLevel>('Начальный')
  const [access, setAccess] = useState<BookAccess>('По подписке')
  const [description, setDescription] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [cover, setCover] = useState<File | null>(null)
  const [error, setError] = useState('')

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (saving) return
    setError('')
    if (!title.trim() || !author.trim() || !technology.trim() || !description.trim() || !file || !cover) {
      setError('Заполните все обязательные поля и выберите файл книги и обложку.')
      return
    }
    if (!file.size || !cover.size) {
      setError('Файл книги и обложка не должны быть пустыми.')
      return
    }
    if (file.size + cover.size >= 100 * 1024 * 1024) {
      setError('Общий размер файлов должен быть меньше 100 МиБ.')
      return
    }
    if (!/\.(pdf|epub)$/i.test(file.name)) {
      setError('Для книги выберите файл PDF или EPUB.')
      return
    }
    if (cover && !/\.(png|jpe?g|webp)$/i.test(cover.name)) {
      setError('Для обложки выберите PNG, JPG или WebP.')
      return
    }
    try {
      await onCreate({ title: title.trim(), author: author.trim(), technology: technology.trim(), level, access, description: description.trim(), file, cover })
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Не удалось загрузить книгу. Попробуйте ещё раз.')
    }
  }

  return <section aria-labelledby="book-form-title" className="pixel-panel mt-6 max-w-4xl p-5 sm:p-8">
    <div className="flex items-start justify-between gap-4"><div><h2 className="text-2xl font-semibold" id="book-form-title">Новая книга</h2><p className="mt-2 text-sm leading-6 text-muted-foreground">Заполните карточку и выберите файл книги и обложку для загрузки на сервер.</p></div><Button aria-label="Закрыть форму" disabled={saving} onClick={onCancel} size="icon" variant="ghost"><X /></Button></div>
    <form className="mt-8 grid gap-6" onSubmit={(event) => void submit(event)}>
      <div className="grid gap-5 sm:grid-cols-2">
        <Field label="Название *"><Input autoFocus className="h-11 border-2 border-border bg-background" maxLength={200} onChange={(event) => setTitle(event.target.value)} placeholder="Например, Чистый код" required value={title} /></Field>
        <Field label="Автор *"><Input className="h-11 border-2 border-border bg-background" maxLength={150} onChange={(event) => setAuthor(event.target.value)} placeholder="Имя автора" required value={author} /></Field>
        <Field label="Тема / технология *"><Input className="h-11 border-2 border-border bg-background" maxLength={100} onChange={(event) => setTechnology(event.target.value)} placeholder="Например, Python" required value={technology} /></Field>
        <Field label="Уровень"><select className="h-11 w-full border-2 border-border bg-background px-3 text-base" onChange={(event) => setLevel(event.target.value as BookLevel)} value={level}><option>Начальный</option><option>Средний</option><option>Продвинутый</option></select></Field>
        <Field label="Доступ"><select className="h-11 w-full border-2 border-border bg-background px-3 text-base" onChange={(event) => setAccess(event.target.value as BookAccess)} value={access}><option>По подписке</option><option>Бесплатно</option></select></Field>
      </div>
      <Field label="Описание *"><Textarea className="min-h-30 border-2 border-border bg-background p-3" maxLength={2000} onChange={(event) => setDescription(event.target.value)} placeholder="Кратко расскажите, о чём эта книга" required value={description} /></Field>
      <div className="grid gap-5 sm:grid-cols-2">
        <FileField accept=".pdf,.epub,application/pdf,application/epub+zip" file={file} hint="PDF или EPUB · обязательное поле" label="Файл книги *" onChange={setFile} />
        <FileField accept=".png,.jpg,.jpeg,.webp,image/png,image/jpeg,image/webp" file={cover} hint="PNG, JPG или WebP · обязательное поле" label="Обложка *" onChange={setCover} />
      </div>
      {error && <p role="alert" className="border-2 border-destructive bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</p>}
      <div className="flex flex-col-reverse gap-3 border-t-2 border-border pt-6 sm:flex-row sm:justify-end"><Button className="h-11 px-5" disabled={saving} onClick={onCancel} type="button" variant="outline">Отмена</Button><Button className="h-11 px-5" disabled={saving} type="submit">{saving ? 'Сохраняем...' : 'Загрузить книгу'}</Button></div>
    </form>
  </section>
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return <label className="grid gap-2 text-sm font-semibold">{label}{children}</label>
}

function FileField({ label, hint, accept, file, onChange }: { label: string; hint: string; accept: string; file: File | null; onChange: (file: File | null) => void }) {
  return <label className="grid gap-2 text-sm font-semibold">{label}<span className="flex min-h-30 cursor-pointer flex-col items-center justify-center gap-2 border-2 border-dashed border-border bg-muted/45 px-4 py-4 text-center hover:bg-muted"><UploadCloud className="size-6" aria-hidden="true" /><span className="max-w-full truncate text-sm font-medium">{file ? file.name : 'Выбрать файл'}</span><span className="text-xs font-normal text-muted-foreground">{hint}</span></span><input accept={accept} className="sr-only" onChange={(event) => onChange(event.target.files?.[0] ?? null)} type="file" /></label>
}
