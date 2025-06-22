import { NextRequest, NextResponse } from 'next/server'
import { ACCESS_TOKEN_KEY, JWT_SECRET } from './shared/constants'
import { jwtVerify } from 'jose'

const secret = new TextEncoder().encode(JWT_SECRET)

async function verifyToken(token?: string) {
  if (!token) return false
  try {
    await jwtVerify(token, secret)
    return true
  } catch {
    return false
  }
}

const PUBLIC_ROUTES = [/^\/login$/]
const PRIVATE_ROUTES = [/^\/profile(\/.*)?$/, /^\/subscriptions(\/.*)?$/]

export async function middleware(request: NextRequest) {
  const token = request.cookies.get(ACCESS_TOKEN_KEY)?.value
  const isAuth = await verifyToken(token)

  const { pathname } = request.nextUrl

  const isPublic = PUBLIC_ROUTES.some((r) => r.test(pathname))
  const isPrivate = PRIVATE_ROUTES.some((r) => r.test(pathname))

  if (isAuth && isPublic) {
    return NextResponse.redirect(new URL('/profile', request.url))
  }

  if (!isAuth && isPrivate) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/login', '/profile/:path*', '/subscriptions/:path*'],
}
