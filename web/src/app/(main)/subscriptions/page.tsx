import { subscriptions } from '@/entities/subscription'
import { SubscriptionCard, SubscriptionModal } from '@/features/subscription'

export default function SubscriptionsPage() {
  return (
    <main className='max-w-5xl w-full mx-auto flex-1'>
      <section className='flex items-center justify-between my-4'>
        <h1 className='text-4xl font-bold'>Мои подписки</h1>
        <SubscriptionModal />
      </section>

      <section className='flex gap-6 flex-wrap py-4 justify-center'>
        {subscriptions.map((item) => (
          <SubscriptionCard key={item.id} item={item} />
        ))}
      </section>
    </main>
  )
}
