'use client'
import { Button } from '@heroui/react'
import Link from 'next/link'
import { FcGoogle } from 'react-icons/fc'
import { FaYandex } from 'react-icons/fa'
import { API_URL } from '@/shared/constants'

export function OAuthButtons() {
  return (
    <div className='flex flex-col gap-2 w-full'>
      <Button
        as={Link}
        href={`${API_URL}/auth/google/login`}
        startContent={<FcGoogle className='size-5' />}
      >
        Продолжить через Google
      </Button>
      <Button
        as={Link}
        href={`${API_URL}/auth/yandex/login`}
        startContent={<FaYandex className='size-5' />}
      >
        Продолжить через Яндекс
      </Button>
    </div>
  )
}
