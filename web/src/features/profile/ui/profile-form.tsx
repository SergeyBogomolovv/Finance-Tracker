'use client'
import { Input } from '@heroui/input'
import { Form } from '@heroui/form'
import { Image } from '@heroui/image'
import { Button } from '@heroui/button'
import type { ChangeEvent, FormEvent } from 'react'
import { addToast } from '@heroui/react'
import { useRef, useState } from 'react'
import { Profile } from '@/entities/profile'
import { S3_BASE_URL } from '@/shared/constants'
import { Divider } from '@heroui/divider'
import { updateProfile } from '../api/update-profile'
import { useRouter } from 'next/navigation'
import { logout } from '@/features/auth'

type Props = {
  profile: Profile
}

export function ProfileForm({ profile }: Props) {
  const router = useRouter()

  const fileInputRef = useRef<HTMLInputElement>(null)
  const [preview, setPreview] = useState(`${S3_BASE_URL}/avatars/${profile.avatar_id}`)

  const providerLabelMap: Record<string, string> = {
    google: 'Google',
    yandex: 'Yandex',
    email: 'Email',
  }

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formEl = e.currentTarget
    const formData = new FormData(formEl)
    updateProfile(formData)
      .then(() => {
        addToast({ title: 'Профиль обновлен' })
      })
      .catch(() => {
        addToast({
          title: 'Возникла непредвиденная ошибка',
          color: 'danger',
        })
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
        defaultValue={profile.full_name ?? ''}
        placeholder='Имя'
        label='Имя'
        labelPlacement='outside'
      />

      <Button type='submit' color='primary' className='w-full'>
        Сохранить
      </Button>

      <Divider />

      <div className='flex flex-col gap-2 w-full'>
        <div className='text-sm text-muted-foreground bg-default-100 rounded-xl p-2.5'>
          Зарегистрирован через: {providerLabelMap[profile.provider] ?? profile.provider}
        </div>
        <div className='text-sm text-muted-foreground bg-default-100 rounded-xl p-2.5'>
          Почта: {profile.email}
        </div>
      </div>
      <Button
        type='button'
        onPress={() => {
          logout().then(() => router.refresh())
        }}
        color='danger'
        className='w-full'
      >
        Выйти
      </Button>
    </Form>
  )
}
