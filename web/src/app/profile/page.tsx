import { fetchCurrentUser } from '@/entities/profile'
import { ProfileForm } from '@/features/profile'
import { redirect } from 'next/navigation'

export default async function ProfilePage() {
  const profile = await fetchCurrentUser()
  if (!profile) {
    redirect('/login')
  }

  return (
    <main className='max-w-5xl w-full mx-auto flex-1 flex flex-col items-center'>
      <section className='mt-10 flex flex-col gap-2'>
        <h1 className='text-4xl font-bold'>Мой профиль</h1>
        <p className='text-muted-foreground'>Здесь вы можете управлять своим профилем.</p>
        <ProfileForm profile={profile} />
      </section>
    </main>
  )
}
