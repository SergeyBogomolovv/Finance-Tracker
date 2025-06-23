'use client'
import { logout } from '@/features/auth'
import { Dropdown, DropdownTrigger, DropdownMenu, DropdownItem } from '@heroui/dropdown'
import { User } from '@heroui/user'
import { useRouter } from 'next/navigation'

export function ProfileDropdown() {
  const router = useRouter()
  return (
    <Dropdown>
      <DropdownTrigger>
        <User
          as='button'
          avatarProps={{
            src: '/icons/yandex_plus.png',
          }}
          className='transition-transform cursor-pointer'
          description='bogomolovs693@gmail.com'
          name='Sergey Bogomolov'
        />
      </DropdownTrigger>
      <DropdownMenu aria-label='Static Actions'>
        <DropdownItem key='profile' href='/profile'>
          Профиль
        </DropdownItem>
        <DropdownItem
          key='delete'
          className='text-danger'
          color='danger'
          onClick={() => {
            logout().then(() => router.refresh())
          }}
        >
          Выйти
        </DropdownItem>
      </DropdownMenu>
    </Dropdown>
  )
}
