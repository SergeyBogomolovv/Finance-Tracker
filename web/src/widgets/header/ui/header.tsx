'use client'
import { Profile } from '@/entities/profile'
import { ProfileDropdown } from '@/features/profile'
import {
  Navbar,
  NavbarBrand,
  NavbarContent,
  NavbarItem,
  NavbarMenu,
  NavbarMenuItem,
  NavbarMenuToggle,
} from '@heroui/navbar'
import { Button, Link } from '@heroui/react'
import { usePathname } from 'next/navigation'
import { useState } from 'react'

type Props = {
  profile: Profile | null
}

export function Header({ profile }: Props) {
  const pathname = usePathname()
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  return (
    <Navbar onMenuOpenChange={setIsMenuOpen}>
      <NavbarContent>
        <NavbarMenuToggle
          aria-label={isMenuOpen ? 'Close menu' : 'Open menu'}
          className='sm:hidden'
        />
        <NavbarBrand>
          <Link href='/' color='foreground' className='font-bold text-inherit'>
            Finance Tracker
          </Link>
        </NavbarBrand>
      </NavbarContent>

      <NavbarContent className='hidden sm:flex gap-4' justify='center'>
        <NavbarItem isActive={pathname === '/subscriptions'}>
          <Link color='foreground' href='/subscriptions'>
            Подписки
          </Link>
        </NavbarItem>
        <NavbarItem isActive={pathname === '/finances'}>
          <Link color='foreground' href='/finances'>
            Финансы
          </Link>
        </NavbarItem>
      </NavbarContent>

      <NavbarContent justify='end'>
        {profile ? (
          <ProfileDropdown profile={profile} />
        ) : (
          <NavbarItem>
            <Button as={Link} color='primary' href='/login' variant='flat'>
              Войти
            </Button>
          </NavbarItem>
        )}
      </NavbarContent>

      <NavbarMenu>
        <NavbarMenuItem isActive={pathname === '/subscriptions'}>
          <Link color='foreground' href='/subscriptions'>
            Подписки
          </Link>
        </NavbarMenuItem>
        <NavbarMenuItem isActive={pathname === '/finances'}>
          <Link color='foreground' href='/finances'>
            Финансы
          </Link>
        </NavbarMenuItem>
      </NavbarMenu>
    </Navbar>
  )
}
