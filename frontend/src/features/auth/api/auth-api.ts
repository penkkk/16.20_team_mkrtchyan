type LoginPayload = {
  login: string
  password: string
}

export type RegisterPayload = {
  email: string
  name: string
  password: string
  surname: string
  tg_username?: string
  username: string
}

export type CurrentUser = {
  id: string
  username: string
  email: string
  tg_username: string
  name: string
  surname: string
}

type LoginResponse = {
  accessToken: string
  expiresIn: number
  refreshExpiresIn: number
  tokenType: string
  user: CurrentUser
}

type RefreshResponse = {
  accessToken: string
}

type ApiErrorResponse = {
  code?: string
  message?: string
}

const API_BASE_URL = '/api/v1'
const ACCESS_TOKEN_STORAGE_KEY = 'itlib.accessToken'
export const currentUserQueryKey = ['current-user'] as const

let accessToken: string | null = null
let refreshPromise: Promise<string | null> | null = null
let sessionVersion = 0

// Earlier frontend versions stored the access token in localStorage.
try {
  localStorage.removeItem(ACCESS_TOKEN_STORAGE_KEY)
} catch {
  // Storage may be unavailable in a private browsing context.
}

async function refreshAccessToken(): Promise<string | null> {
  if (refreshPromise) return refreshPromise

  const startedAtVersion = sessionVersion
  refreshPromise = (async () => {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, { credentials: 'include' })
    if (response.status === 401) {
      if (sessionVersion === startedAtVersion) accessToken = null
      return accessToken
    }
    if (!response.ok) throw new Error('Не удалось восстановить сессию.')

    const result = (await response.json()) as RefreshResponse
    if (sessionVersion === startedAtVersion) accessToken = result.accessToken
    return accessToken
  })().finally(() => { refreshPromise = null })

  return refreshPromise
}

async function requestCurrentUser(token: string): Promise<Response> {
  return fetch(`${API_BASE_URL}/me`, {
    credentials: 'include',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function getCurrentUser(): Promise<CurrentUser | null> {
  let token = accessToken ?? await refreshAccessToken()
  if (!token) return null

  let response = await requestCurrentUser(token)
  if (response.status === 401) {
    if (accessToken === token) accessToken = null
    token = await refreshAccessToken()
    if (!token) return null
    response = await requestCurrentUser(token)
  }

  if (response.status === 401) {
    if (accessToken === token) accessToken = null
    return null
  }
  if (!response.ok) throw new Error('Не удалось загрузить профиль.')
  return response.json() as Promise<CurrentUser>
}

export async function logout(): Promise<void> {
  if (refreshPromise) await refreshPromise.catch(() => null)
  const response = await fetch(`${API_BASE_URL}/auth/logout`, {
    method: 'DELETE',
    credentials: 'include',
    headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : {},
  })
  if (!response.ok && response.status !== 401) throw new Error('Не удалось выйти из аккаунта.')
  sessionVersion += 1
  accessToken = null
}

export async function login(payload: LoginPayload): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE_URL}/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    const error = (await response.json().catch(() => null)) as ApiErrorResponse | null
    throw new Error(error?.message ?? 'Не удалось войти. Проверьте логин и пароль.')
  }

  const result = (await response.json()) as LoginResponse
  sessionVersion += 1
  accessToken = result.accessToken
  return result
}

export async function registerUser(payload: RegisterPayload): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE_URL}/auth/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    const error = (await response.json().catch(() => null)) as ApiErrorResponse | null
    throw new Error(error?.message ?? 'Не удалось зарегистрироваться.')
  }

  const result = (await response.json()) as LoginResponse
  sessionVersion += 1
  accessToken = result.accessToken
  return result
}

export function getOAuthStartUrl(provider: 'google' | 'yandex') {
  const url = new URL(`${API_BASE_URL}/auth/${provider}/start`, window.location.origin)

  url.searchParams.set('return_url', '/profile')

  return url.toString()
}
