'use client'
import {
  Input,
  Button,
  Form,
  Card,
  CardHeader,
  CardBody,
  CardFooter,
  Divider,
  addToast,
} from '@heroui/react'
import { useState, type FormEvent } from 'react'
import { InputOtp } from '@heroui/input-otp'
import Link from 'next/link'
import { FcGoogle } from 'react-icons/fc'
import { FaYandex } from 'react-icons/fa'
import { API_URL } from '@/shared/constants'
import { requestEmailCode, verifyEmailCode } from '@/features/auth/api/email'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'

type Props = {
  error?: string
}

export function AuthForm({ error }: Props) {
  const [email, setEmail] = useState('')
  const [isCodeSent, setCodeSent] = useState(false)
  const [loading, setLoading] = useState(false)
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const clearErrorQueryParam = () => {
    if (!searchParams) return
    if (!searchParams.has('error')) return
    const next = new URLSearchParams(searchParams.toString())
    next.delete('error')
    const query = next.toString()
    const href = query ? `${pathname}?${query}` : pathname
    router.replace(href)
  }

  const sendCode = async (emailToSend: string) => {
    setLoading(true)
    try {
      const response = await requestEmailCode(emailToSend)

      if (response.ok) {
        clearErrorQueryParam()
        addToast({ title: 'Код отправлен вам на почту' })
        setCodeSent(true)
        return
      }

      if (response.status === 400) {
        addToast({ title: 'Попробуйте войти другим способом', color: 'danger' })
        return
      }
      if (response.status === 429) {
        addToast({ title: 'Превышен лимит. Попробуйте позже', color: 'danger' })
        return
      }
      addToast({ title: 'Что-то пошло не так', color: 'danger' })
    } catch {
      addToast({ title: 'Нет соединения. Проверьте сеть', color: 'danger' })
    } finally {
      setLoading(false)
    }
  }

  const emailAuth = async (otp: string) => {
    setLoading(true)
    try {
      const response = await verifyEmailCode(email, otp)
      if (response.ok) {
        router.refresh()
        return
      }
      if (response.status === 401) {
        addToast({ title: 'Неверный код' })
        return
      }
      addToast({ title: 'Что-то пошло не так' })
    } catch {
      addToast({ title: 'Нет соединения. Проверьте сеть' })
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    const data = Object.fromEntries(formData)

    if (!isCodeSent) {
      const emailFromForm = data.email as string
      setEmail(emailFromForm)
      await sendCode(emailFromForm)
    } else {
      await emailAuth((data.otp as string) || '')
    }
  }

  return (
    <Card className='w-[350px] rounded-4xl p-4'>
      <CardHeader className='flex flex-col gap-2 items-center'>
        <h2 className='text-lg font-semibold'>Finance Tracker</h2>
        <p className='text-content4-foreground text-center text-sm'>
          Войдите или зарегистрируйтесь, чтобы управлять своими расходами и подписками.
        </p>
      </CardHeader>
      <CardBody>
        <Form className='flex flex-col gap-4 items-center' onSubmit={handleSubmit}>
          <Input
            isRequired
            label='Ваша почта'
            labelPlacement='outside'
            type='email'
            placeholder='example@example.com'
            name='email'
            isDisabled={isCodeSent || loading}
          />

          {isCodeSent && (
            <InputOtp
              description='Введите код из письма'
              autoFocus
              isRequired
              length={6}
              name='otp'
              size='md'
              isDisabled={loading}
            />
          )}

          <Button type='submit' color='primary' className='w-full' isLoading={loading}>
            {isCodeSent ? 'Подтвердить' : 'Отправить код'}
          </Button>
        </Form>
      </CardBody>

      <Divider />

      <CardFooter className='flex flex-col gap-3 items-center'>
        {error && (
          <div className='bg-red-600 opacity-60 text-sm p-2 rounded-lg w-full text-center text-white'>
            {error === 'oauth_failed' ? 'Попробуйте войти другим способом' : error}
          </div>
        )}

        <Button
          className='w-full'
          as={Link}
          href={`${API_URL}/auth/google/login`}
          startContent={<FcGoogle className='size-5' />}
        >
          Продолжить через Google
        </Button>
        <Button
          className='w-full'
          as={Link}
          href={`${API_URL}/auth/yandex/login`}
          startContent={<FaYandex className='size-5' />}
        >
          Продолжить через Яндекс
        </Button>
      </CardFooter>
    </Card>
  )
}
