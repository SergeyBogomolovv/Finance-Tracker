'use client'
import { Button, Input, Select, SelectItem, Switch, Form, addToast } from '@heroui/react'
import { Subscription } from '@/entities/subscription'
import { DateInput } from '@heroui/react'
import { fromUnixTime } from 'date-fns'
import { CalendarDate } from '@internationalized/date'
import { useRouter } from 'next/navigation'
import { FormEvent, useMemo, useState } from 'react'
import Link from 'next/link'

export function SubscriptionForm({ sub }: { sub: Subscription }) {
  const router = useRouter()
  const [isAutoPay, setIsAutoPay] = useState(sub.is_autopay)
  const [isNotify, setIsNotify] = useState(sub.notify)

  const nextPaymentDate = useMemo(() => {
    const date = fromUnixTime(sub.next_payment_date)
    return new CalendarDate(date.getFullYear(), date.getMonth() + 1, date.getDate())
  }, [sub.next_payment_date])

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const data = {
      ...Object.fromEntries(new FormData(e.currentTarget)),
      is_autopay: isAutoPay,
      notify: isNotify,
    }
    addToast({
      title: 'Информация обновлена',
    })
    console.log(data)
  }

  return (
    <Form onSubmit={handleSubmit}>
      <div className='flex flex-col gap-3 p-4 md:p-0'>
        <h1 className='border-0 ring-0 bg-transparent outline-0 w-full sm:text-4xl text-3xl font-bold'>
          {sub.name}
        </h1>
        <p className='border-0 ring-0 bg-transparent outline-0 w-full'>{sub.notes}</p>

        <Switch name='notify' isSelected={isNotify} onValueChange={setIsNotify}>
          <div className='flex flex-col gap-1'>
            <p className='text-medium'>Включить уведомления</p>
            <p className='text-tiny text-default-400'>
              Они будут приходить к вам на почту за день до списания
            </p>
          </div>
        </Switch>

        <Switch name='is_autopay' isSelected={isAutoPay} onValueChange={setIsAutoPay}>
          <div className='flex flex-col gap-1'>
            <p className='text-medium'>Автоматический платеж</p>
            <p className='text-tiny text-default-400'>
              Включите если платеж происходит автоматически
            </p>
          </div>
        </Switch>

        <Button as={Link} href={sub.link || '/'} target='_blank' variant='flat' color='primary'>
          Перейти
        </Button>

        <Input
          name='amount'
          label='Стоимость'
          defaultValue={String(sub.amount)}
          placeholder='0.00'
          startContent={
            <div className='pointer-events-none flex items-center'>
              <span className='text-default-400 text-small'>
                {sub.currency === 'RUB' ? '₽' : '$'}
              </span>
            </div>
          }
          type='number'
        />

        <Select
          name='currency'
          label='Валюта'
          placeholder='Выберите валюту'
          defaultSelectedKeys={[sub.currency]}
          radius='md'
        >
          <SelectItem key='USD'>USD</SelectItem>
          <SelectItem key='RUB'>RUB</SelectItem>
        </Select>

        <Select
          name='status'
          label='Статус'
          placeholder='Выберите статус'
          defaultSelectedKeys={[sub.status]}
          radius='md'
        >
          <SelectItem key='active'>Активна</SelectItem>
          <SelectItem key='cancelled'>Отменена</SelectItem>
          <SelectItem key='trial'>Пробный период</SelectItem>
          <SelectItem key='expired'>Истекла</SelectItem>
        </Select>

        <DateInput
          name='next_payment_date'
          label='Дата следующего платежа'
          defaultValue={nextPaymentDate}
        />

        <Select
          name='frequency'
          label='Частота оплаты'
          placeholder='Выберите частоту'
          defaultSelectedKeys={[sub.frequency]}
          radius='md'
        >
          <SelectItem key='year'>Год</SelectItem>
          <SelectItem key='half_year'>6 Месяцев</SelectItem>
          <SelectItem key='quarter'>3 Месяца</SelectItem>
          <SelectItem key='month'>Месяц</SelectItem>
          <SelectItem key='week'>Неделя</SelectItem>
          <SelectItem key='once'>Единоразово</SelectItem>
        </Select>

        <Button type='submit' color='primary'>
          Сохранить
        </Button>
        <Button type='button' onPress={() => router.back()} variant='ghost'>
          Назад
        </Button>
      </div>
    </Form>
  )
}
