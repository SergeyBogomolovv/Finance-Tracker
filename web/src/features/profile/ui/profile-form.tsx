'use client'
import { Input } from '@heroui/input'
import { Form } from '@heroui/form'
import { Image } from '@heroui/image'
import { Button } from '@heroui/button'
import type { ChangeEvent, FormEvent } from 'react'
import { addToast } from '@heroui/react'
import { useRef, useState } from 'react'

export function ProfileForm() {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [preview, setPreview] = useState('/icons/yandex_plus.png')

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const data = Object.fromEntries(new FormData(e.currentTarget))
    console.log(data)
    addToast({
      title: 'Профиль обновлен',
    })
  }

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = () => {
        if (typeof reader.result === 'string') {
          setPreview(reader.result)
        }
      }
      reader.readAsDataURL(file)
    }
  }

  return (
    <Form onSubmit={handleSubmit} className='w-full max-w-md space-y-2 mt-2'>
      <div className='relative overflow-hidden group'>
        <Image src={preview} alt='Profile icon' width={200} height={200} />
        <input
          type='file'
          name='avatar'
          accept='image/*'
          ref={fileInputRef}
          onChange={handleFileChange}
          hidden
        />
        <div
          className='absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-80 text-white font-medium text-sm bg-black bg-opacity-40 transition-opacity z-10 cursor-pointer'
          onClick={() => fileInputRef.current?.click()}
        >
          Изменить фото
        </div>
      </div>

      <Input
        name='name'
        value={'Сергей Богомолов'}
        placeholder='Имя'
        label='Имя'
        labelPlacement='outside'
      />

      <Input
        name='email'
        value={'bogomolovs693@email.com'}
        placeholder='email@email.com'
        label='Почта'
        labelPlacement='outside'
        description='На этот адрес будут приходить уведомления'
      />

      <Button type='submit' color='primary'>
        Сохранить
      </Button>
    </Form>
  )
}
