import { z } from 'zod'

export const registerSchema = z
  .object({
    name: z.string().min(2, 'Минимум 2 символа'),
    email: z.string().email('Неверный email'),
    password: z.string().min(6, 'Минимум 6 символов'),
    passwordRepeat: z.string().min(6, 'Минимум 6 символов'),
  })
  .superRefine(({ passwordRepeat, password }, ctx) => {
    if (passwordRepeat !== password) {
      ctx.addIssue({
        code: 'custom',
        message: 'Пароли не совпадают',
        path: ['passwordRepeat'],
      })
    }
  })

export type RegisterSchema = z.infer<typeof registerSchema>
