'use client'
import { Subscription } from '@/entities/subscription'
import { Card, CardHeader, CardBody, CardFooter, Divider, Link, Image } from '@heroui/react'
import { format, isTomorrow, isToday, addDays, isSameDay } from 'date-fns'
import { ru } from 'date-fns/locale'

export function SubscriptionCard({ item }: { item: Subscription }) {
  return (
    <Card className='max-w-[400px] w-full lg:w-fit'>
      <CardHeader className='flex flex-col gap-2 items-start'>
        <h4 className='text-lg font-semibold'>{item.name}</h4>
        <p className='text-muted-foreground'>{item.notes}</p>
      </CardHeader>
      <CardBody className='items-center'>
        <Image
          src={`/icons/${item.service}.png`}
          alt={item.name}
          width={300}
          height={300}
          className='mx-auto w-full object-contain'
        />
      </CardBody>
      <Divider />
      <CardFooter className='flex flex-col gap-2 items-start'>
        <p>
          {item.amount} {item.currency}
        </p>
        <p>Следующий платеж: {formatDate(item.next_payment_date)}</p>
        <Link showAnchorIcon href={`/subscriptions/${item.id}`}>
          Подробнее
        </Link>
      </CardFooter>
    </Card>
  )
}

function formatDate(unix: number): string {
  const date = new Date(unix * 1000)
  const now = new Date()

  if (isToday(date)) {
    return 'сегодня'
  }

  if (isTomorrow(date)) {
    return 'завтра'
  }

  if (isSameDay(date, addDays(now, 2))) {
    return 'послезавтра'
  }

  if (date.getFullYear() === now.getFullYear()) {
    return format(date, 'd MMMM', { locale: ru })
  }

  return format(date, 'dd.MM.yyyy')
}
