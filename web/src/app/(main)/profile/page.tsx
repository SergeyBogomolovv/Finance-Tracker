import { ProfileForm } from '@/features/profile'

export default function ProfilePage() {
  return (
    <main className='max-w-5xl w-full mx-auto flex-1 flex justify-center'>
      <section className='mt-10 flex flex-col gap-2'>
        <h1 className='text-4xl font-bold'>Мой профиль</h1>
        <p className='text-muted-foreground'>Здесь вы можете управлять своим профилем.</p>
        <ProfileForm />
      </section>
    </main>
  )
}
