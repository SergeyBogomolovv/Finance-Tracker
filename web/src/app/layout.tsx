import type { Metadata } from 'next'
import { Geist, Geist_Mono } from 'next/font/google'
import { Providers } from './providers'
import { Header } from '@/widgets/header'
import Image from 'next/image'
import './globals.css'
import { checkAuth } from '@/shared/utils/auth'

const geistSans = Geist({
  variable: '--font-geist-sans',
  subsets: ['latin'],
})

const geistMono = Geist_Mono({
  variable: '--font-geist-mono',
  subsets: ['latin'],
})

export const metadata: Metadata = {
  title: 'Finance Tracker',
  description: 'Сервис для отслеживания финансов',
}

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  const isAuth = await checkAuth()
  return (
    <html lang='ru' className='dark'>
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        <Providers>
          <div className='min-h-screen flex flex-col'>
            <div className='fixed inset-0 -z-10' aria-hidden='true'>
              <Image
                src='/background.jpeg'
                alt='Background'
                fill
                className='blur-sm brightness-60 object-center object-cover'
              />
            </div>
            <Header isAuthenticated={isAuth} />
            {children}
          </div>
        </Providers>
      </body>
    </html>
  )
}
