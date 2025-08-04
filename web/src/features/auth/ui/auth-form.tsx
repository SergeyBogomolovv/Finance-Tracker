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
import { useRouter } from 'next/navigation'
import { sendOTP } from '../api/send-otp'
import { verifyOTP } from '../api/verify-otp'

export function AuthForm() {
  const [email, setEmail] = useState('')
  const [isCodeSent, setCodeSent] = useState(false)
  const [loading, setLoading] = useState(false)
  const router = useRouter()

  const sendCode = async (email: string) => {
    setLoading(true)
    await sendOTP(email)
    addToast({ title: 'Код отправлен' })
    setCodeSent(true)
    setLoading(false)
  }

  const emailAuth = async (otp: string) => {
    setLoading(true)
    await verifyOTP(email, otp)
    addToast({ title: 'Аутентификация прошла' })
    setLoading(false)
    router.refresh()
  }

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    const data = Object.fromEntries(formData)

    if (!isCodeSent) {
      setEmail(data.email as string)
      await sendCode(data.email as string)
    } else {
      await emailAuth(data.otp as string)
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
