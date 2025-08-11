'use client'
import { Profile } from '@/entities/profile'
import { logout } from '@/features/auth'
import { Dropdown, DropdownTrigger, DropdownMenu, DropdownItem } from '@heroui/dropdown'
import { User } from '@heroui/user'
import { useRouter } from 'next/navigation'
import { MEDIA_URL } from '@/shared/constants'

type Props = {
  profile: Profile
}

export function ProfileDropdown({ profile }: Props) {
  const router = useRouter()
  return (
    <Dropdown>
      <DropdownTrigger>
        <User
          as='button'
          avatarProps={{
            src: `${MEDIA_URL}/${profile.avatar_id}`,
          }}
          className='transition-transform cursor-pointer'
          classNames={{
            name: 'hidden md:inline',
            description: 'hidden md:inline',
          }}
          description={profile.email}
          name={profile.full_name || 'Unknown'}
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
