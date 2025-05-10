import Image from 'next/image'
import { Button } from '@heroui/button'
import { FcGoogle } from 'react-icons/fc'
import { RegisterForm } from '@/features/auth'

export default function Home() {
  return (
    <main className='flex min-h-screen flex-col md:flex-row'>
      {/* Левая часть */}
      <div className='relative hidden md:flex w-full md:w-1/2 items-center justify-center bg-muted p-8'>
        <div className='text-center max-w-md'>
          <Image
            src='/register.svg'
            alt='Register'
            width={400}
            height={400}
            className='mx-auto mb-6'
          />
          <h1 className='text-3xl font-bold'>Finance Tracker</h1>
          <p className='text-muted-foreground mt-2'>Контролируй расходы и подписки легко.</p>
        </div>
      </div>

      {/* Правая часть */}
      <div className='w-full md:w-1/2 flex items-center justify-center px-6 py-12'>
        <div className='w-full max-w-md space-y-6'>
          <div>
            <h2 className='text-2xl font-bold text-center'>Регистрация</h2>
          </div>
          <RegisterForm />
          <Button className='w-full' startContent={<FcGoogle className='size-5' />}>
            Продолжить через Google
          </Button>
        </div>
      </div>
    </main>
  )
}
