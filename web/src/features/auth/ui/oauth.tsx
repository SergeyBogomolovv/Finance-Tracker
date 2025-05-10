'use client'
import { Button } from '@heroui/react'
import Link from 'next/link'
import { FcGoogle } from 'react-icons/fc'

export function OAuthButton() {
  return (
    <Link href='/auth/google' className='w-full'>
      <Button className='w-full' startContent={<FcGoogle className='size-5' />}>
        Продолжить через Google
      </Button>
    </Link>
  )
}
