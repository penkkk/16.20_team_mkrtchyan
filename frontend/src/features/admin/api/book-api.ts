import { authenticatedFetch } from '@/features/auth/api/auth-api'
import type { BookAccess, BookLevel } from '../model/book-store'

export type BookUpload = {
  title: string
  author: string
  description: string
  technology: string
  level: BookLevel
  access: BookAccess
  file: File
  cover: File
}

const levels: Record<BookLevel, string> = {
  'Начальный': 'beginner',
  'Средний': 'intermediate',
  'Продвинутый': 'advanced',
}

const accessLevels: Record<BookAccess, string> = {
  'Бесплатно': 'free',
  'По подписке': 'subscription',
}

export async function uploadBook(book: BookUpload): Promise<void> {
  const form = new FormData()
  form.append('title', book.title)
  form.append('author', book.author)
  form.append('description', book.description)
  form.append('topic', book.technology)
  form.append('level', levels[book.level])
  form.append('access_level', accessLevels[book.access])
  form.append('book', book.file)
  form.append('cover', book.cover)

  let response: Response
  try {
    response = await authenticatedFetch('/admin/upload', { method: 'POST', body: form })
  } catch (error) {
    if (error instanceof TypeError) throw new Error('Не удалось связаться с сервером. Проверьте соединение.', { cause: error })
    throw error
  }
  if (response.ok) return
  if (response.status === 403) throw new Error('Загружать книги может только администратор.')
  if (response.status === 400 || response.status === 413) throw new Error('Проверьте поля и содержимое файлов: книга должна быть PDF или EPUB, обложка — PNG, JPEG или WebP. Размер запроса — до 100 МиБ.')
  throw new Error('Не удалось загрузить книгу на сервер. Попробуйте ещё раз.')
}

export type CatalogBook = {
  id: string
  title: string
  author: string
  description: string
  topic: string
  access_level: 'free' | 'subscription'
  level: 'beginner' | 'intermediate' | 'advanced'
}

export type CatalogPage = {
  books: CatalogBook[]
  limit: number
  offset: number
  has_more: boolean
}

export async function listBooks(offset: number, search: string, topic: string): Promise<CatalogPage> {
  const params = new URLSearchParams({ limit: '20', offset: String(offset), search, topic })
  let response: Response
  try {
    response = await authenticatedFetch(`/admin/books?${params}`)
  } catch (error) {
    if (error instanceof TypeError) throw new Error('Не удалось связаться с сервером. Проверьте соединение.', { cause: error })
    throw error
  }
  if (response.status === 403) throw new Error('Просматривать каталог может только администратор.')
  if (!response.ok) throw new Error('Не удалось загрузить каталог. Попробуйте ещё раз.')
  return response.json() as Promise<CatalogPage>
}

async function bookRequest(id: string, operation: 'download' | 'delete'): Promise<Response> {
  let response: Response
  try {
    const path = `/admin/books/${encodeURIComponent(id)}${operation === 'download' ? '/download' : ''}`
    response = await authenticatedFetch(path, { method: operation === 'delete' ? 'DELETE' : 'GET' })
  } catch (error) {
    if (error instanceof TypeError) throw new Error('Не удалось связаться с сервером. Проверьте соединение.', { cause: error })
    throw error
  }
  if (response.status === 403) throw new Error('Это действие доступно только администратору.')
  if (response.status === 404) throw new Error('Книга не найдена. Обновите каталог.')
  if (!response.ok) throw new Error(operation === 'delete' ? 'Не удалось удалить книгу. Попробуйте ещё раз.' : 'Не удалось скачать книгу. Попробуйте ещё раз.')
  return response
}

export async function deleteBook(id: string): Promise<void> {
  await bookRequest(id, 'delete')
}

export async function downloadBook(id: string): Promise<void> {
  const response = await bookRequest(id, 'download')
  const disposition = response.headers.get('Content-Disposition') ?? ''
  const encoded = disposition.match(/filename\*=utf-8''([^;]+)/i)?.[1]
  const plain = disposition.match(/filename=(?:"([^";]+)"|([^;]+))/i)
  let filename = plain?.[1] || plain?.[2]?.trim() || 'book'
  if (encoded) {
    try { filename = decodeURIComponent(encoded) } catch { /* use the fallback filename */ }
  }
  const url = URL.createObjectURL(await response.blob())
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), 1000)
}
