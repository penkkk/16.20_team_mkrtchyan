import { z } from 'zod'

export const registerSchema = z.object({
  name: z.string().trim().min(1, 'Введите имя'),
  surname: z.string().trim().min(1, 'Введите фамилию'),
  username: z
    .string()
    .trim()
    .min(3, 'Минимум 3 символа')
    .regex(/^[a-zA-Z0-9_]+$/, 'Только латиница, цифры и _'),
  email: z.email('Введите корректный email'),
  password: z.string().min(8, 'Минимум 8 символов'),
  tg_username: z.string().trim().optional(),
})

export type RegisterFormValues = z.input<typeof registerSchema>
