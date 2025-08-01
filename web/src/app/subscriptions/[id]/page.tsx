import { subscriptions } from '@/entities/subscription'
import { SubscriptionForm } from '@/widgets/subscription'
import Image from 'next/image'

export default function SubscriptionPage({ params }: { params: { id: string } }) {
  const sub = subscriptions.find((s) => s.id === Number(params.id))!
  return (
    <main className='max-w-5xl w-full mx-auto flex-1'>
      <section className='flex gap-10 justify-center items-start mt-12'>
        <Image
          src={`/icons/${sub.service}.png`}
          alt={sub.name}
          width={200}
          height={200}
          className='rounded-lg hidden md:block'
        />
        <SubscriptionForm sub={sub} />
      </section>
    </main>
  )
}
