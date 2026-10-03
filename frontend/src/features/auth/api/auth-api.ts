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

type LoginResponse = {
  accessToken: string
  expiresIn: number
  refreshExpiresIn: number
  tokenType: string
  user: {
    id: string
    username: string
    email: string
    tg_username: string
    name: string
    surname: string
  }
}

type ApiErrorResponse = {
  code?: string
  message?: string
}

const API_BASE_URL = '/api/v1'
const ACCESS_TOKEN_STORAGE_KEY = 'itlib.accessToken'

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

  return response.json() as Promise<LoginResponse>
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

  return response.json() as Promise<LoginResponse>
}

export function persistAccessToken(accessToken: string) {
  localStorage.setItem(ACCESS_TOKEN_STORAGE_KEY, accessToken)
}

export function getOAuthStartUrl(provider: 'google' | 'yandex') {
  const url = new URL(`${API_BASE_URL}/auth/${provider}/start`, window.location.origin)

  url.searchParams.set('return_url', '/profile')

  return url.toString()
}
