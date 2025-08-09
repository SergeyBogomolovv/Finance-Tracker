import { API_URL } from '@/shared/constants'

export async function requestEmailCode(email: string): Promise<Response> {
  return fetch(`${API_URL}/auth/email`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email }),
  })
}

export async function verifyEmailCode(email: string, otp: string): Promise<Response> {
  return fetch(`${API_URL}/auth/email/verify`, {
    credentials: 'include',
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, otp }),
  })
}
