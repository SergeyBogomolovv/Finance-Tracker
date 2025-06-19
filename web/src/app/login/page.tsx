import Image from 'next/image'
import Link from 'next/link'
import { LoginForm, OAuthButtons } from '@/features/auth'

export default function LoginPage() {
  return (
    <main className='flex min-h-screen flex-col md:flex-row'>
      <section className='hidden md:flex w-full flex-1 items-center justify-center p-8 bg-gray-800'>
        <div className='text-center max-w-md'>
          <Image src='/login.svg' alt='Login' width={400} height={400} className='mx-auto mb-6' />
          <h1 className='text-3xl font-bold'>С возвращением</h1>
          <p className='text-muted-foreground mt-2'>
            Войдите в свой аккаунт, чтобы управлять своими расходами и подписками.
          </p>
        </div>
      </section>

      <section className='w-full flex-1 flex items-center justify-center px-6 py-12'>
        <div className='w-full max-w-md space-y-6 flex flex-col items-center'>
          <h2 className='text-2xl font-bold'>Вход</h2>
          <LoginForm />
          <OAuthButtons />
          <Link href='/register' className='text-sm hover:underline'>
            У вас еще нет аккаунта? Регистрация
          </Link>
        </div>
      </section>
    </main>
  )
}
