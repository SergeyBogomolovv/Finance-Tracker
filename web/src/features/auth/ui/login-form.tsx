'use client'

import { useState } from 'react'
import { Input, Button, Form } from '@heroui/react'
import { loginSchema } from '../model/login-schema'
import type { FormEvent } from 'react'

export function LoginForm() {
  const [errors, setErrors] = useState<Record<string, string[]>>({})

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const data = Object.fromEntries(new FormData(e.currentTarget))
    const result = loginSchema.safeParse(data)

    if (!result.success) {
      setErrors(result.error.flatten().fieldErrors)
      return
    }

    setErrors({})
    console.log('Вход выполнен:', result.data)
  }

  return (
    <Form className='w-full max-w-md space-y-4' validationErrors={errors} onSubmit={handleSubmit}>
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

      <Button color='primary' className='w-full' type='submit'>
        Войти
      </Button>
    </Form>
  )
}
