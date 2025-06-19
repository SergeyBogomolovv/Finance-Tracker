import Image from 'next/image'
import Link from 'next/link'
import { OAuthButtons, RegisterForm } from '@/features/auth'

export default function RegisterPage() {
  return (
    <main className='flex min-h-screen flex-col md:flex-row'>
      <section className='hidden md:flex w-full flex-1 items-center justify-center p-8 bg-gray-800'>
        <div className='text-center max-w-md'>
          <Image
            src='/register.svg'
            alt='Register'
            width={400}
            height={400}
            className='mx-auto mb-6'
          />
          <h1 className='text-3xl font-bold'>Finance Tracker</h1>
          <p className='text-muted-foreground mt-2'>
            Контроль над подписками и финансами начинается здесь – управляйте бюджетом легко и
            уверенно.
          </p>
        </div>
      </section>

      <section className='w-full flex-1 flex items-center justify-center px-6 py-12'>
        <div className='w-full max-w-md space-y-6 flex flex-col items-center'>
          <h2 className='text-2xl font-bold'>Регистрация</h2>
          <RegisterForm />
          <OAuthButtons />
          <Link href='/login' className='text-sm hover:underline'>
            Уже есть аккаунт? Вход
          </Link>
        </div>
      </section>
    </main>
  )
}
