'use server'
import { cookies } from 'next/headers'
import { ACCESS_TOKEN_KEY, API_URL } from '@/shared/constants'
import { revalidateTag } from 'next/cache'

export async function updateProfile(formData: FormData): Promise<void> {
  const cookieStore = await cookies()
  const token = cookieStore.get(ACCESS_TOKEN_KEY)?.value
  if (!token) throw new Error('Нет токена авторизации')

  const res = await fetch(`${API_URL}/profile/update`, {
    method: 'PUT',
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: formData,
  })

  if (!res.ok) {
    throw new Error('failed to update profile')
  }
  revalidateTag('profile')
}
