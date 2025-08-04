import { API_URL } from '@/shared/constants'

export async function sendOTP(email: string) {
  await fetch(`${API_URL}/auth/email`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email }),
  })
}
