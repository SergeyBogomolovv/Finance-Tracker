'use client'

import { useState } from 'react'
import { Input, Button, Form } from '@heroui/react'
import { registerSchema } from '../model/register-schema' // путь адаптируй
import type { FormEvent } from 'react'

export function RegisterForm() {
  const [errors, setErrors] = useState<Record<string, string[]>>({})

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const data = Object.fromEntries(new FormData(e.currentTarget))
    const result = registerSchema.safeParse(data)

    if (!result.success) {
      setErrors(result.error.flatten().fieldErrors)
      return
    }

    setErrors({})
    console.log('Регистрация успешна:', result.data)
  }

  return (
    <Form className='w-full max-w-md space-y-4' validationErrors={errors} onSubmit={handleSubmit}>
      <Input
        isRequired
        label='Имя'
        labelPlacement='outside'
        placeholder='Иван Иванов'
        name='name'
        errorMessage={errors.name?.join(', ')}
      />

      <Input
        isRequired
        label='Email'
        labelPlacement='outside'
        type='email'
        placeholder='example@example.com'
        name='email'
        errorMessage={errors.email?.join(', ')}
      />

      <Input
        isRequired
        label='Пароль'
        labelPlacement='outside'
        type='password'
        placeholder='******'
        name='password'
        errorMessage={errors.password?.join(', ')}
      />

      <Input
        isRequired
        label='Подтвердите пароль'
        labelPlacement='outside'
        type='password'
        placeholder='******'
        name='passwordRepeat'
        errorMessage={errors.passwordRepeat?.join(', ')}
      />

      <Button color='primary' className='w-full' type='submit'>
        Зарегистрироваться
      </Button>
    </Form>
  )
}
