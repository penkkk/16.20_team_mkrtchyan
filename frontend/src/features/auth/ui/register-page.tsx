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
import { registerUser, persistAccessToken, type RegisterPayload } from '@/features/auth/api/auth-api'
import { registerSchema, type RegisterFormValues } from '@/features/auth/model/auth-schemas'
import { cn } from '@/lib/utils'
import { AuthLayout } from './auth-layout'

export function RegisterPage() {
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
          <h1 className="max-w-lg text-5xl leading-tight font-medium text-(--itlib-ice)">
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
