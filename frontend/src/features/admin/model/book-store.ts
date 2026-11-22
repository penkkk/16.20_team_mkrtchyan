export type BookLevel = 'Начальный' | 'Средний' | 'Продвинутый'
export type BookAccess = 'По подписке' | 'Бесплатно'

export type BookRecord = {
  id: string
  title: string
  author: string
  technology: string
  level: BookLevel
  access: BookAccess
  description: string
  fileName: string
  file: File | null
  cover: File | null
  createdAt: string
  isExample: boolean
}

const databaseName = 'itlib-admin-demo'
const storeName = 'books'

const examples: BookRecord[] = [
  {
    id: 'example-clean-code', title: 'Чистый код', author: 'Роберт Мартин',
    technology: 'Разработка', level: 'Средний', access: 'По подписке',
    description: 'Приёмы написания понятного и поддерживаемого кода.',
    fileName: '', file: null, cover: null, createdAt: '2026-11-01T12:00:00.000Z', isExample: true,
  },
  {
    id: 'example-algorithms', title: 'Грокаем алгоритмы', author: 'Адитья Бхаргава',
    technology: 'Алгоритмы', level: 'Начальный', access: 'По подписке',
    description: 'Наглядное введение в алгоритмы и структуры данных.',
    fileName: '', file: null, cover: null, createdAt: '2026-11-02T12:00:00.000Z', isExample: true,
  },
  {
    id: 'example-refactoring', title: 'Рефакторинг', author: 'Мартин Фаулер',
    technology: 'Разработка', level: 'Продвинутый', access: 'По подписке',
    description: 'Каталог приёмов улучшения существующего кода.',
    fileName: '', file: null, cover: null, createdAt: '2026-11-03T12:00:00.000Z', isExample: true,
  },
]

function openDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(databaseName, 1)
    request.onupgradeneeded = () => {
      const database = request.result
      const store = database.createObjectStore(storeName, { keyPath: 'id' })
      examples.forEach((book) => store.put(book))
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

export async function listBooks(): Promise<BookRecord[]> {
  const database = await openDatabase()
  try {
    return await new Promise<BookRecord[]>((resolve, reject) => {
      const request = database.transaction(storeName, 'readonly').objectStore(storeName).getAll()
      request.onsuccess = () => resolve((request.result as BookRecord[]).sort((a, b) => b.createdAt.localeCompare(a.createdAt)))
      request.onerror = () => reject(request.error)
    })
  } finally {
    database.close()
  }
}

export async function addBook(book: Omit<BookRecord, 'id' | 'createdAt' | 'isExample'>): Promise<void> {
  const database = await openDatabase()
  try {
    await new Promise<void>((resolve, reject) => {
      const transaction = database.transaction(storeName, 'readwrite')
      transaction.objectStore(storeName).add({
        ...book,
        id: crypto.randomUUID(),
        createdAt: new Date().toISOString(),
        isExample: false,
      } satisfies BookRecord)
      transaction.oncomplete = () => resolve()
      transaction.onerror = () => reject(transaction.error)
      transaction.onabort = () => reject(transaction.error)
    })
  } finally {
    database.close()
  }
}

export async function deleteBook(id: string): Promise<void> {
  const database = await openDatabase()
  try {
    await new Promise<void>((resolve, reject) => {
      const transaction = database.transaction(storeName, 'readwrite')
      transaction.objectStore(storeName).delete(id)
      transaction.oncomplete = () => resolve()
      transaction.onerror = () => reject(transaction.error)
      transaction.onabort = () => reject(transaction.error)
    })
  } finally {
    database.close()
  }
}
