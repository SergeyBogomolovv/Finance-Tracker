import { profileSchema } from '../model/profile-schema'
import { API_URL } from '@/shared/constants'
import { cookies } from 'next/headers'

export async function fetchCurrentUser() {
  try {
    const cookiesStore = await cookies()
    const token = cookiesStore.get('access_token')?.value
    if (!token) return null

    const res = await fetch(`${API_URL}/profile/me`, {
      headers: { Authorization: `Bearer ${token}` },
    })

    if (!res.ok) return null

    const data = await res.json()
    return profileSchema.parse(data)
  } catch (error) {
    return null
  }
}
