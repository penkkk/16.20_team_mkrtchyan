import { Link, useNavigate } from '@tanstack/react-router'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import {
  ArrowRight,
  AtSign,
  BookOpenText,
  LockKeyhole,
  Mail,
  Sparkles,
  User,
  type LucideIcon,
} from 'lucide-react'
import {
  useForm,
  type FieldError,
  type SubmitHandler,
  type UseFormRegisterReturn,
} from 'react-hook-form'
import { Button } from '@/components/ui/button'
import { registerUser, persistAccessToken, type RegisterPayload, getOAuthStartUrl } from '@/features/auth/api/auth-api'
import { registerSchema, type RegisterFormValues } from '@/features/auth/model/auth-schemas'
import { cn } from '@/lib/utils'
import { AuthLayout } from './auth-layout'
import { useState } from 'react'

type OAuthProvider = 'google' | 'yandex'

const oauthProviders: Array<{
  label: string
  provider: OAuthProvider
  mark: React.JSX.Element;
}> = [
  { label: 'Продолжить с Google', provider: 'google', mark: (<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24"><path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/><path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/><path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/><path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/><path d="M1 1h22v22H1z" fill="none"/></svg>)},
  { label: 'Продолжить с Яндекс', provider: 'yandex', mark: (<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M2.04 12c0-5.523 4.476-10 10-10 5.522 0 10 4.477 10 10s-4.478 10-10 10c-5.524 0-10-4.477-10-10z" fill="#FC3F1D"/><path d="M13.32 7.666h-.924c-1.694 0-2.585.858-2.585 2.123 0 1.43.616 2.1 1.881 2.959l1.045.704-3.003 4.487H7.49l2.695-4.014c-1.55-1.111-2.42-2.19-2.42-4.015 0-2.288 1.595-3.85 4.62-3.85h3.003v11.868H13.32V7.666z" fill="#fff"/></svg>)},
]

export function RegisterPage() {
  const [showForm, setShowForm] = useState(false);

  function handleOAuthRegister(provider: 'google' | 'yandex') {
    window.location.assign(getOAuthStartUrl(provider))
  }



  const navigate = useNavigate()
  const form = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: '',
      name: '',
      password: '',
      surname: '',
      tg_username: '',
      username: '',
    },
  })

  const registerMutation = useMutation({
    mutationFn: registerUser,
    onSuccess: (result) => {
      persistAccessToken(result.accessToken)
      void navigate({ to: '/profile' })
    },
  })

  const submitRegister: SubmitHandler<RegisterFormValues> = (values) => {
    const payload: RegisterPayload = {
      email: values.email.trim(),
      name: values.name.trim(),
      password: values.password,
      surname: values.surname.trim(),
      username: values.username.trim(),
    }

    const tgUsername = values.tg_username?.trim()
    if (tgUsername) {
      payload.tg_username = tgUsername
    }

    registerMutation.mutate(payload)
  }

  const requestError =
    registerMutation.error instanceof Error
      ? registerMutation.error.message
      : registerMutation.isError
        ? 'Не удалось зарегистрироваться.'
        : null

  return (
    <AuthLayout>
      <section className="grid w-full max-w-5xl items-center gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(390px,460px)]">
        <div className="hidden max-w-xl lg:block">
          <div className="mb-8 inline-flex items-center gap-2 rounded-full border border-border bg-muted px-3 py-1 text-sm text-muted-foreground">
            <Sparkles className="size-4 text-(--itlib-ice)" aria-hidden="true" />
            ITLib
          </div>
          <h1 className="glass-heading max-w-lg text-5xl leading-tight font-medium">
            Создайте аккаунт и соберите свою учебную библиотеку.
          </h1>
          <p className="mt-5 max-w-md text-base leading-7 text-muted-foreground">
            Регистрация откроет профиль, сохраненные материалы и персональные возможности ITLib.
          </p>
        </div>

        <div className="glass-auth-card w-full max-w-[460px] justify-self-end rounded-2xl border p-6 text-card-foreground sm:p-8">
          <div className="mb-8 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl border border-border bg-muted">
              <BookOpenText className="size-5 text-(--itlib-ice)" aria-hidden="true" />
            </div>
            <div>
              <p className="text-sm font-medium tracking-[0.14em] text-muted-foreground uppercase">
                ITLib
              </p>
              <h2 className="text-2xl font-medium text-foreground">Создать аккаунт</h2>
            </div>
          </div>

          {showForm ? (
                      <form className="grid gap-4" onSubmit={form.handleSubmit(submitRegister)}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField
                autoComplete="given-name"
                error={form.formState.errors.name}
                icon={User}
                label="Имя"
                placeholder="Иван"
                registration={form.register('name')}
              />
              <FormField
                autoComplete="family-name"
                error={form.formState.errors.surname}
                icon={User}
                label="Фамилия"
                placeholder="Иванов"
                registration={form.register('surname')}
              />
            </div>

            <FormField
              autoComplete="username"
              error={form.formState.errors.username}
              icon={AtSign}
              label="Логин"
              placeholder="ivan_2026"
              registration={form.register('username')}
            />

            <FormField
              autoComplete="email"
              error={form.formState.errors.email}
              icon={Mail}
              label="Email"
              placeholder="name@example.com"
              registration={form.register('email')}
              type="email"
            />

            <FormField
              autoComplete="off"
              error={form.formState.errors.tg_username}
              icon={AtSign}
              label="Telegram"
              placeholder="username"
              registration={form.register('tg_username')}
            />

            <FormField
              autoComplete="new-password"
              error={form.formState.errors.password}
              icon={LockKeyhole}
              label="Пароль"
              placeholder="Минимум 8 символов"
              registration={form.register('password')}
              type="password"
            />

            {requestError ? (
              <p className="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {requestError}
              </p>
            ) : null}

            <Button className="mt-1 h-11 rounded-md text-base" disabled={registerMutation.isPending} type="submit">
              {registerMutation.isPending ? 'Создаем...' : 'Создать аккаунт'}
              <ArrowRight className="size-4" aria-hidden="true" />
            </Button>
          </form>
          ) : (

            <div className="grid gap-3 sm:grid-cols-1 my-6">
              <p className="mb-2 text-sm text-muted-foreground">
                Выберите удобный способ создания аккаунта
              </p>
              {oauthProviders.map(({ label, provider, mark }) => (
                <Button
                  className="h-11 justify-start gap-3 border-border bg-muted/70 px-3 text-foreground hover:bg-accent"
                  key={provider}
                  type="button"
                  variant="outline"
                  onClick={() => handleOAuthRegister(provider)}
                >
                  
                  <span className="grid size-6 place-items-center">
                    {mark}
                  </span>
                  {label}
                </Button>
              ))}
                  <Button
                    className="h-11 mt-5"
                    type="button"
                    onClick={() => setShowForm(true)}
                  >
                    Создать аккаунт по email
                  </Button>
            </div>
            
           )}
          <p className="mt-6 text-center text-sm text-muted-foreground">
            Уже есть аккаунт?{' '}
            <Link className="font-medium text-foreground underline underline-offset-4" to="/login">
              Войти
            </Link>
          </p>
        </div>
      </section>
    </AuthLayout>
  )
}

type FormFieldProps = {
  autoComplete: string
  error?: FieldError
  icon: LucideIcon
  label: string
  placeholder: string
  registration: UseFormRegisterReturn
  type?: 'email' | 'password' | 'text'
}

function FormField({
  autoComplete,
  error,
  icon: Icon,
  label,
  placeholder,
  registration,
  type = 'text',
}: FormFieldProps) {
  const inputId = registration.name

  return (
    <label className="grid gap-2 text-sm font-medium text-foreground" htmlFor={inputId}>
      {label}
      <span className="relative">
        <Icon
          className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
          aria-hidden="true"
        />
        <input
          aria-invalid={Boolean(error)}
          autoComplete={autoComplete}
          className={cn(
            'h-11 w-full rounded-md border border-border bg-input px-10 text-base text-foreground outline-none transition',
            'placeholder:text-muted-foreground/70 focus:border-ring focus:ring-3 focus:ring-ring/30',
            error && 'border-destructive/70 focus:border-destructive focus:ring-destructive/20',
          )}
          id={inputId}
          placeholder={placeholder}
          type={type}
          {...registration}
        />
      </span>
      {error ? <span className="text-xs font-normal text-destructive">{error.message}</span> : null}
    </label>
  )
}

