import { cookies } from 'next/headers'
import { jwtVerify } from 'jose'
import { ACCESS_TOKEN_KEY, JWT_SECRET } from '../constants'

const secret = new TextEncoder().encode(JWT_SECRET)

export async function checkAuth(): Promise<boolean> {
  const cookieStore = await cookies()
  const token = cookieStore.get(ACCESS_TOKEN_KEY)?.value

  if (!token) return false

  try {
    await jwtVerify(token, secret)
    return true
  } catch {
    return false
  }
}
