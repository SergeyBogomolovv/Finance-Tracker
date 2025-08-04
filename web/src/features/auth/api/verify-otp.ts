import { API_URL } from '@/shared/constants'

export async function verifyOTP(email: string, otp: string) {
  const res = await fetch(`${API_URL}/auth/email/verify`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, otp }),
  })
  return res.json()
}
