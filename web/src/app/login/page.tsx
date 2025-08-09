import { AuthForm } from '@/features/auth'

type Props = {
  searchParams: Promise<{ error?: string }>
}

export default async function LoginPage({ searchParams }: Props) {
  const params = await searchParams

  return (
    <main className='flex-1 flex items-center justify-center'>
      <AuthForm error={params.error} />
    </main>
  )
}
