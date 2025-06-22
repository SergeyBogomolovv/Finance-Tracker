import { Button } from '@heroui/button'
import { Link } from '@heroui/link'
import Image from 'next/image'

export default function Home() {
  return (
    <main className='w-full max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-16 space-y-32'>
      {/* Hero Section */}
      <section className='flex flex-col-reverse md:flex-row items-center justify-between gap-10'>
        <div className='flex-1 flex flex-col justify-center gap-6'>
          <h1 className='text-4xl sm:text-5xl font-bold leading-tight'>
            Финансовая ясность начинается здесь.
          </h1>
          <p className='text-lg max-w-md'>
            Finance Tracker помогает тебе отслеживать расходы, управлять подписками и находить
            способы экономить.
          </p>
          <Button
            as={Link}
            href='/subscriptions'
            className='w-fit px-6 py-3 text-base'
            color='primary'
          >
            Управлять финансами
          </Button>
        </div>
        <div className='flex-1 w-full max-w-md md:max-w-lg'>
          <Image
            src='/landing/hero.svg'
            alt='hero'
            width={500}
            height={500}
            className='w-full h-auto'
            priority
          />
        </div>
      </section>

      {/* Problem Section */}
      <section className='flex flex-col-reverse md:flex-row items-center gap-10'>
        <div className='flex-1 max-w-md md:max-w-lg'>
          <Image
            src='/landing/receipt.svg'
            alt='problem'
            width={500}
            height={500}
            className='w-full h-auto'
            priority
          />
        </div>
        <div className='flex-1 space-y-4'>
          <h2 className='text-3xl font-bold'>Не знаешь, куда уходят деньги?</h2>
          <ul className='list-disc pl-5 space-y-2'>
            <li>Списания за подписки, о которых ты забыл</li>
            <li>Неконтролируемые расходы</li>
            <li>Нет полной картины твоего бюджета</li>
          </ul>
          <p className='font-bold'>Finance Tracker решает всё это за тебя.</p>
        </div>
      </section>

      {/* Features Section */}
      <section className='text-center space-y-12'>
        <h2 className='text-3xl font-bold'>Что ты получаешь</h2>
        <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8'>
          {[
            {
              title: 'Гибкое управление подписками',
              desc: 'Настраивай напоминания и статусы — платная, пробная, отменена.',
              icon: '/landing/setup.svg',
            },
            {
              title: 'Категоризация расходов',
              desc: 'Все траты сортируются по категориям: еда, транспорт, развлечения и т.д.',
              icon: '/landing/categorization.svg',
            },
            {
              title: 'Уведомления о списаниях',
              desc: 'Уведомим заранее, если приближается списание.',
              icon: '/landing/notifications.svg',
            },
            {
              title: 'История и фильтрация трат',
              desc: 'Просматривай всю историю и фильтруй по дате, категории, подписке или сумме.',
              icon: '/landing/history.svg',
            },
            {
              title: 'Финансовые цели',
              desc: 'Создавай цели и следи за их выполнением.',
              icon: '/landing/checklist.svg',
            },
            {
              title: 'Безопасность и приватность',
              desc: 'Никакой лишней интеграции — только ты решаешь, что и как отслеживать.',
              icon: '/landing/secure.svg',
            },
          ].map((feature, i) => (
            <div key={i} className='flex flex-col items-center gap-4 text-center'>
              <Image
                src={feature.icon}
                alt={feature.title}
                width={128}
                height={128}
                className='size-22'
              />
              <h3 className='font-semibold text-lg'>{feature.title}</h3>
              <p className='text-muted-foreground text-sm'>{feature.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Testimonials Section */}
      <section className='space-y-10 text-center'>
        <h2 className='text-3xl font-bold'>Отзывы пользователей</h2>
        <div className='grid gap-8 sm:grid-cols-2 lg:grid-cols-3'>
          {[
            {
              name: 'Алексей, 28',
              text: 'Я впервые за долгое время понимаю, куда уходят мои деньги.',
            },
            {
              name: 'Ольга, 33',
              text: 'Удалила 5 подписок за ненадобностью. Экономия — 3 000₽ в месяц!',
            },
            {
              name: 'Ирина, 24',
              text: 'Теперь трачу осознанно. Приложение стало ежедневной привычкой.',
            },
          ].map((review, i) => (
            <div key={i} className='bg-white p-6 rounded-2xl shadow-md text-left space-y-3'>
              <p className='text-black'>&ldquo;{review.text}&rdquo;</p>
              <p className='text-sm text-black'>— {review.name}</p>
            </div>
          ))}
        </div>
      </section>

      {/* CTA Section */}
      <section className='bg-primary text-white py-16 px-6 rounded-3xl text-center space-y-6'>
        <h2 className='text-3xl font-bold'>Начни контролировать финансы уже сегодня</h2>
        <p className='text-lg max-w-xl mx-auto'>
          Подключи свои счета, настрой уведомления и забудь о неожиданных списаниях.
        </p>
        <Button as={Link} href='/subscriptions' className='px-8 py-4 text-base font-semibold'>
          Начать бесплатно
        </Button>
      </section>
    </main>
  )
}
