'use server'
import { ACCESS_TOKEN_KEY } from '@/shared/constants'
import { cookies } from 'next/headers'

export async function logout() {
  const cookieStore = await cookies()
  cookieStore.delete(ACCESS_TOKEN_KEY)
}
