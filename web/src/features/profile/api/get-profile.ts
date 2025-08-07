import { API_URL } from '@/shared/constants'
import { cookies } from 'next/headers'

export async function getProfile() {
  const cookiesStore = await cookies()

  const res = await fetch(`${API_URL}/profile/me`, {
    headers: {
      Authorization: `Bearer ${cookiesStore.get('access_token')?.value}`,
    },
  })

  return res.json()
}
