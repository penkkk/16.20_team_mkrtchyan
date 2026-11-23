import { useCallback, useEffect, useMemo, useRef, useState, type FormEvent } from 'react'
import { Link } from '@tanstack/react-router'
import { BookOpenText, Download, FilePlus2, FileText, LibraryBig, Search, Trash2, UploadCloud, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ThemeToggle } from '@/components/theme-toggle'
import { addBook, deleteBook, listBooks, type BookAccess, type BookLevel, type BookRecord } from '../model/book-store'

type View = 'books' | 'create'

export function AdminPage() {
  const [view, setView] = useState<View>('books')
  const [books, setBooks] = useState<BookRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState('Все темы')
  const [bookToDelete, setBookToDelete] = useState<BookRecord | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [notice, setNotice] = useState('')

  const reload = useCallback(async () => {
    try {
      const nextBooks = await listBooks()
      setBooks(nextBooks)
      setLoadError(false)
    } catch {
      setLoadError(true)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    let active = true
    void listBooks()
      .then((nextBooks) => {
        if (active) setBooks(nextBooks)
      })
      .catch(() => {
        if (active) setLoadError(true)
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => { active = false }
  }, [])

  const technologies = useMemo(() => ['Все темы', ...new Set(books.map((book) => book.technology))], [books])
  const visibleBooks = useMemo(() => {
    const query = search.trim().toLocaleLowerCase('ru')
    return books.filter((book) =>
      (filter === 'Все темы' || book.technology === filter) &&
      (!query || `${book.title} ${book.author} ${book.technology}`.toLocaleLowerCase('ru').includes(query)),
    )
  }, [books, filter, search])

  async function handleDelete() {
    if (!bookToDelete) return
    setDeleting(true)
    try {
      await deleteBook(bookToDelete.id)
      setNotice(`Книга «${bookToDelete.title}» удалена из локального каталога.`)
      setBookToDelete(null)
      await reload()
    } catch {
      setNotice('Не удалось удалить книгу. Попробуйте ещё раз.')
    } finally {
      setDeleting(false)
    }
  }

  async function handleCreate(book: Parameters<typeof addBook>[0]) {
    await addBook(book)
    await reload()
    setView('books')
    setNotice(`Книга «${book.title}» добавлена в локальный каталог.`)
  }

  return (
    <div className="relative isolate min-h-svh bg-background text-foreground">
      <div aria-hidden="true" className="home-background pointer-events-none absolute inset-0 -z-10" />
      <header className="border-b-2 border-border bg-card/95">
        <div className="mx-auto flex min-h-18 max-w-7xl items-center justify-between gap-4 px-5 sm:px-8">
          <div className="flex items-center gap-3">
            <Link aria-label="На главную ITLib" className="pixel-frame grid size-10 place-items-center text-foreground" to="/">
              <BookOpenText className="size-5" aria-hidden="true" />
            </Link>
            <div>
              <p className="text-sm font-semibold tracking-wide">ITLib</p>
              <p className="text-sm text-muted-foreground">Управление библиотекой</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <span className="hidden border border-border bg-muted px-3 py-1.5 text-sm text-foreground sm:inline-block">Демо-режим</span>
            <ThemeToggle />
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-7xl px-5 pb-20 pt-8 sm:px-8 sm:pt-12">
        <div className="flex flex-col justify-between gap-5 sm:flex-row sm:items-end">
          <div>
            <p className="mb-3 text-sm font-semibold uppercase tracking-[0.16em] text-muted-foreground">Панель управления</p>
            <h1 className="pixel-heading text-3xl font-semibold sm:text-4xl">Каталог книг</h1>
            <p className="mt-4 max-w-2xl text-base leading-7 text-muted-foreground">
              Добавляйте книги, просматривайте список и удаляйте записи.
            </p>
          </div>
          {view === 'books' && (
            <Button className="h-11 gap-2 px-5 text-base" onClick={() => { setNotice(''); setView('create') }}>
              <FilePlus2 className="size-5" aria-hidden="true" /> Добавить книгу
            </Button>
          )}
        </div>

        <div className="mt-8 border-2 border-border bg-secondary px-5 py-4 text-sm leading-6 text-secondary-foreground sm:flex sm:items-center sm:gap-3">
          <span className="mb-1 inline-block font-semibold sm:mb-0">Локальное демо</span>
          <span>Книги и выбранные файлы сохраняются только в этом браузере. С сервером каталог пока не связан; вход администратора ещё не проверяется.</span>
        </div>

        {notice && (
          <div role="status" className="mt-5 flex items-start justify-between gap-3 border-2 border-border bg-card px-4 py-3 text-sm">
            <span>{notice}</span>
            <button aria-label="Скрыть сообщение" className="text-muted-foreground hover:text-foreground" onClick={() => setNotice('')} type="button"><X className="size-4" /></button>
          </div>
        )}

        <div className="mt-8 grid gap-4 sm:grid-cols-3">
          <Stat label="Всего книг" value={books.length} icon={<LibraryBig className="size-5" />} />
          <Stat label="Загружено локально" value={books.filter((book) => !book.isExample).length} icon={<UploadCloud className="size-5" />} />
          <Stat label="Примеры без файлов" value={books.filter((book) => book.isExample).length} icon={<FileText className="size-5" />} />
        </div>

        <div className="mt-9 flex gap-2 border-b-2 border-border" aria-label="Разделы админки">
          <button aria-current={view === 'books' ? 'page' : undefined} className={`px-4 py-3 text-sm font-semibold ${view === 'books' ? 'border-b-4 border-primary text-foreground' : 'text-muted-foreground hover:text-foreground'}`} onClick={() => setView('books')} type="button">Все книги</button>
          <button aria-current={view === 'create' ? 'page' : undefined} className={`px-4 py-3 text-sm font-semibold ${view === 'create' ? 'border-b-4 border-primary text-foreground' : 'text-muted-foreground hover:text-foreground'}`} onClick={() => setView('create')} type="button">Добавить книгу</button>
        </div>

        {view === 'books' ? (
          <section aria-label="Список книг" className="pixel-panel mt-6 p-5 sm:p-7">
            <div className="flex flex-col justify-between gap-4 lg:flex-row lg:items-center">
              <div>
                <h2 className="text-xl font-semibold">Все книги</h2>
                <p className="mt-1 text-sm text-muted-foreground">В этом браузере: {books.length}</p>
              </div>
              <div className="flex flex-col gap-3 sm:flex-row">
                <label className="relative block sm:w-70">
                  <span className="sr-only">Поиск по книгам</span>
                  <Search aria-hidden="true" className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                  <Input className="h-11 border-2 border-border bg-background pl-10" onChange={(event) => setSearch(event.target.value)} placeholder="Название или автор" value={search} />
                </label>
                <label>
                  <span className="sr-only">Фильтр по теме</span>
                  <select className="h-11 min-w-42 border-2 border-border bg-background px-3 text-sm text-foreground" onChange={(event) => setFilter(event.target.value)} value={filter}>
                    {technologies.map((technology) => <option key={technology}>{technology}</option>)}
                  </select>
                </label>
              </div>
            </div>

            {loadError ? (
              <div className="py-16 text-center"><p>Не удалось открыть локальный каталог.</p><Button className="mt-4" onClick={() => void reload()}>Повторить</Button></div>
            ) : loading ? (
              <p className="py-16 text-center text-muted-foreground">Загружаем книги...</p>
            ) : visibleBooks.length === 0 ? (
              <div className="py-16 text-center">
                <BookOpenText className="mx-auto size-10 text-muted-foreground" aria-hidden="true" />
                <h3 className="mt-4 text-lg font-semibold">{books.length ? 'Книги не найдены' : 'В каталоге пока нет книг'}</h3>
                <p className="mt-2 text-sm text-muted-foreground">{books.length ? 'Попробуйте изменить поиск или фильтр.' : 'Добавьте первую книгу через форму.'}</p>
                {!books.length && <Button className="mt-5" onClick={() => setView('create')}>Добавить книгу</Button>}
              </div>
            ) : (
              <div className="mt-6 overflow-x-auto">
                <table className="w-full min-w-185 border-collapse text-left text-sm">
                  <thead><tr className="border-y-2 border-border bg-muted text-muted-foreground">
                    <th className="px-4 py-3 font-semibold">Книга</th><th className="px-4 py-3 font-semibold">Тема / уровень</th><th className="px-4 py-3 font-semibold">Доступ</th><th className="px-4 py-3 font-semibold">Файл</th><th className="px-4 py-3 text-right font-semibold">Действия</th>
                  </tr></thead>
                  <tbody>{visibleBooks.map((book) => <BookRow book={book} key={book.id} onDelete={() => setBookToDelete(book)} />)}</tbody>
                </table>
              </div>
            )}
          </section>
        ) : <BookForm onCancel={() => setView('books')} onCreate={handleCreate} />}
      </main>

      {bookToDelete && <DeleteDialog book={bookToDelete} deleting={deleting} onCancel={() => setBookToDelete(null)} onConfirm={() => void handleDelete()} />}
    </div>
  )
}

function Stat({ label, value, icon }: { label: string; value: number; icon: React.ReactNode }) {
  return <div className="pixel-panel flex items-center justify-between gap-4 px-5 py-5"><div><p className="text-sm text-muted-foreground">{label}</p><p className="mt-2 text-3xl font-semibold">{value}</p></div><span className="grid size-11 place-items-center border-2 border-border bg-muted text-foreground">{icon}</span></div>
}

function BookRow({ book, onDelete }: { book: BookRecord; onDelete: () => void }) {
  function download() {
    if (!book.file) return
    const url = URL.createObjectURL(book.file)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = book.fileName
    anchor.click()
    window.setTimeout(() => URL.revokeObjectURL(url), 1000)
  }

  return <tr className="border-b border-border/70 align-middle hover:bg-muted/45">
    <td className="px-4 py-4"><div className="flex min-w-55 items-center gap-3"><BookCover cover={book.cover} /><div><p className="font-semibold text-foreground">{book.title}</p><p className="mt-1 text-muted-foreground">{book.author}</p></div></div></td>
    <td className="px-4 py-4"><p>{book.technology}</p><p className="mt-1 text-muted-foreground">{book.level}</p></td>
    <td className="px-4 py-4"><span className="inline-block border border-border bg-secondary px-2 py-1 text-xs font-medium">{book.access}</span></td>
    <td className="px-4 py-4">{book.file ? <button className="flex max-w-45 items-center gap-2 text-left text-foreground underline underline-offset-4 hover:text-primary" onClick={download} title={book.fileName} type="button"><Download className="size-4 shrink-0" aria-hidden="true" /><span className="truncate">{book.fileName}</span></button> : <span className="text-muted-foreground">Пример без файла</span>}</td>
    <td className="px-4 py-4 text-right"><Button aria-label={`Удалить книгу «${book.title}»`} className="size-9" onClick={onDelete} size="icon" title="Удалить книгу" variant="destructive"><Trash2 className="size-4" aria-hidden="true" /></Button></td>
  </tr>
}

function BookCover({ cover }: { cover: File | null }) {
  const [url, setUrl] = useState<string | null>(null)
  useEffect(() => {
    if (!cover) return
    const objectUrl = URL.createObjectURL(cover)
    const image = new Image()
    image.onload = () => setUrl(objectUrl)
    image.src = objectUrl
    return () => {
      image.onload = null
      URL.revokeObjectURL(objectUrl)
    }
  }, [cover])

  return <span className="grid size-12 shrink-0 place-items-center overflow-hidden border-2 border-border bg-secondary">{url ? <img alt="" className="size-full object-cover" src={url} /> : <BookOpenText className="size-5" aria-hidden="true" />}</span>
}

function BookForm({ onCancel, onCreate }: { onCancel: () => void; onCreate: (book: Parameters<typeof addBook>[0]) => Promise<void> }) {
  const [title, setTitle] = useState('')
  const [author, setAuthor] = useState('')
  const [technology, setTechnology] = useState('')
  const [level, setLevel] = useState<BookLevel>('Начальный')
  const [access, setAccess] = useState<BookAccess>('По подписке')
  const [description, setDescription] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [cover, setCover] = useState<File | null>(null)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    if (!title.trim() || !author.trim() || !technology.trim() || !description.trim() || !file) {
      setError('Заполните все обязательные поля и выберите файл книги.')
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
    setSaving(true)
    try {
      await onCreate({ title: title.trim(), author: author.trim(), technology: technology.trim(), level, access, description: description.trim(), fileName: file.name, file, cover })
    } catch {
      setError('Не удалось сохранить книгу в браузере. Проверьте свободное место и попробуйте ещё раз.')
    } finally {
      setSaving(false)
    }
  }

  return <section aria-labelledby="book-form-title" className="pixel-panel mt-6 max-w-4xl p-5 sm:p-8">
    <div className="flex items-start justify-between gap-4"><div><h2 className="text-2xl font-semibold" id="book-form-title">Новая книга</h2><p className="mt-2 text-sm leading-6 text-muted-foreground">Заполните карточку и выберите PDF или EPUB. Файл останется в этом браузере.</p></div><Button aria-label="Закрыть форму" onClick={onCancel} size="icon" variant="ghost"><X /></Button></div>
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
        <FileField accept=".png,.jpg,.jpeg,.webp,image/png,image/jpeg,image/webp" file={cover} hint="PNG, JPG или WebP · необязательно" label="Обложка" onChange={setCover} />
      </div>
      {error && <p role="alert" className="border-2 border-destructive bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</p>}
      <div className="flex flex-col-reverse gap-3 border-t-2 border-border pt-6 sm:flex-row sm:justify-end"><Button className="h-11 px-5" onClick={onCancel} type="button" variant="outline">Отмена</Button><Button className="h-11 px-5" disabled={saving} type="submit">{saving ? 'Сохраняем...' : 'Сохранить книгу локально'}</Button></div>
    </form>
  </section>
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return <label className="grid gap-2 text-sm font-semibold">{label}{children}</label>
}

function FileField({ label, hint, accept, file, onChange }: { label: string; hint: string; accept: string; file: File | null; onChange: (file: File | null) => void }) {
  return <label className="grid gap-2 text-sm font-semibold">{label}<span className="flex min-h-30 cursor-pointer flex-col items-center justify-center gap-2 border-2 border-dashed border-border bg-muted/45 px-4 py-4 text-center hover:bg-muted"><UploadCloud className="size-6" aria-hidden="true" /><span className="max-w-full truncate text-sm font-medium">{file ? file.name : 'Выбрать файл'}</span><span className="text-xs font-normal text-muted-foreground">{hint}</span></span><input accept={accept} className="sr-only" onChange={(event) => onChange(event.target.files?.[0] ?? null)} type="file" /></label>
}

function DeleteDialog({ book, deleting, onCancel, onConfirm }: { book: BookRecord; deleting: boolean; onCancel: () => void; onConfirm: () => void }) {
  const cancelRef = useRef<HTMLButtonElement>(null)
  useEffect(() => {
    cancelRef.current?.focus()
    const onKeyDown = (event: KeyboardEvent) => { if (event.key === 'Escape' && !deleting) onCancel() }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [deleting, onCancel])

  return <div className="fixed inset-0 z-50 grid place-items-center bg-black/60 px-4"><div aria-describedby="delete-description" aria-labelledby="delete-title" aria-modal="true" className="pixel-panel w-full max-w-md bg-card p-6 sm:p-8" role="alertdialog"><span className="grid size-12 place-items-center border-2 border-border bg-destructive/10 text-destructive"><Trash2 className="size-6" /></span><h2 className="mt-5 text-xl font-semibold" id="delete-title">Удалить книгу?</h2><p className="mt-3 text-sm leading-6 text-muted-foreground" id="delete-description">«{book.title}» будет удалена из локального каталога{book.file ? ' вместе с сохранённым файлом' : ''}. Отменить действие нельзя.</p><div className="mt-7 flex justify-end gap-3"><Button onClick={onCancel} ref={cancelRef} type="button" variant="outline">Отмена</Button><Button disabled={deleting} onClick={onConfirm} type="button" variant="destructive">{deleting ? 'Удаляем...' : 'Удалить'}</Button></div></div></div>
}
