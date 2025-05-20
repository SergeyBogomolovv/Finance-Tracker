'use client'
import {
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  useDisclosure,
} from '@heroui/modal'
import { Button } from '@heroui/button'
import { FaPlus } from 'react-icons/fa'
import { Form } from '@heroui/form'
import { DateInput, Input, Select, SelectItem, Switch, Textarea } from '@heroui/react'
import { FormEvent, useState } from 'react'

export function SubscriptionModal() {
  const { isOpen, onOpen, onOpenChange } = useDisclosure()
  const [isAutoPay, setIsAutoPay] = useState(false)
  const [isNotify, setIsNotify] = useState(false)

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const data = {
      ...Object.fromEntries(new FormData(e.currentTarget)),
      is_autopay: isAutoPay,
      notify: isNotify,
    }
    console.log(data)
  }

  return (
    <>
      <Button startContent={<FaPlus />} color='primary' onPress={onOpen}>
        Добавить
      </Button>
      <Modal backdrop='blur' isDismissable={false} isOpen={isOpen} onOpenChange={onOpenChange}>
        <ModalContent>
          {(onClose) => (
            <Form onSubmit={handleSubmit}>
              <ModalHeader className='flex flex-col gap-1'>Новая подписка</ModalHeader>
              <ModalBody>
                <Input isRequired label='Название' name='name' />
                <Textarea label='Заметки' name='notes' />

                <Select
                  isRequired
                  name='currency'
                  label='Валюта'
                  placeholder='Выберите валюту'
                  radius='md'
                >
                  <SelectItem key='USD'>USD</SelectItem>
                  <SelectItem key='RUB'>RUB</SelectItem>
                </Select>

                <Input
                  isRequired
                  name='amount'
                  label='Стоимость'
                  placeholder='0.00'
                  startContent={
                    <div className='pointer-events-none flex items-center'>
                      <span className='text-default-400 text-small'>$</span>
                    </div>
                  }
                  type='number'
                />

                <DateInput isRequired name='next_payment_date' label='Дата следующего платежа' />

                <Select
                  isRequired
                  name='frequency'
                  label='Частота оплаты'
                  placeholder='Выберите частоту'
                  radius='md'
                >
                  <SelectItem key='year'>Год</SelectItem>
                  <SelectItem key='half_year'>6 Месяцев</SelectItem>
                  <SelectItem key='quarter'>3 Месяца</SelectItem>
                  <SelectItem key='month'>Месяц</SelectItem>
                  <SelectItem key='week'>Неделя</SelectItem>
                  <SelectItem key='once'>Единоразово</SelectItem>
                </Select>

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
              </ModalBody>
              <ModalFooter>
                <Button type='button' color='danger' variant='light' onPress={onClose}>
                  Отмена
                </Button>
                <Button type='submit' color='primary'>
                  Добавить
                </Button>
              </ModalFooter>
            </Form>
          )}
        </ModalContent>
      </Modal>
    </>
  )
}
