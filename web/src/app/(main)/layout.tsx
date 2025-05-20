import { Header } from '@/widgets/header'

export default function MainLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <div className='min-h-screen flex flex-col'>
      <Header isAuthenticated={true} />

      {children}
    </div>
  )
}
