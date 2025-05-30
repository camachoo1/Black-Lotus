'use client';

import { useState, useEffect } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { AuthModal } from '../Modals/auth';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';

export function Navbar() {
  const [isSticky, setIsSticky] = useState(false); // handle navbar sticking to top
  const pathname = usePathname();
  const { user, logout } = useAuth();
  const isHome = pathname === '/';

  // on scroll navbar will stick and display additional info
  useEffect(() => {
    const handleScroll = () => {
      setIsSticky(window.scrollY > 100);
    };

    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const showFullNav = !isHome || isSticky;

  const handleLogout = async () => {
    await logout();
  };

  return (
    <nav
      className={`fixed top-0 left-0 w-full z-50 bg-white transition-all duration-300 ${
        showFullNav ? 'shadow' : 'shadow-none'
      }`}
    >
      <div className='max-w-screen-xl mx-auto flex justify-between items-center px-6 py-4'>
        {/* LEFT: Logo */}
        <Link href='/'>
          <div className='flex items-center gap-2'>
            <Image
              src='/filled.svg'
              alt='logo'
              width={40}
              height={40}
            />
            <h1 className='text-lg font-bold custom-text'>
              Black Lotus
            </h1>
          </div>
        </Link>

        {/* CENTER: Links + Search */}
        <div
          className={`flex items-center justify-between gap-8 flex-grow mx-8 transition-all duration-500 ${
            showFullNav
              ? 'opacity-100 translate-y-0'
              : 'opacity-0 -translate-y-2 pointer-events-none'
          }`}
        >
          {/* LEFT side links */}
          <div className='flex gap-6'>
            <Link
              href='/'
              className='custom-text text-sm font-semibold text-gray-900 hover:text-cyan-500'
            >
              Home
            </Link>
            <Link
              href='/hotels'
              className='custom-text text-sm font-semibold text-gray-900 hover:text-cyan-500'
            >
              Hotels
            </Link>
          </div>

          {/* RIGHT side search bar */}
          <div className='flex items-center border rounded-lg px-3 py-2 text-gray-500 w-[240px]'>
            <span className='mr-2'>🔍</span>
            <input
              type='text'
              placeholder='Explore by destination'
              className='custom-text w-full text-sm outline-none'
            />
          </div>
        </div>

        {/* RIGHT: Auth buttons */}
        {user ? (
          <>
            <Link href='/account'>
              <button className='custom-text text-xs px-5 py-2.5 font-semibold rounded-lg hover:bg-gray-100 text-gray-900 hover:cursor-pointer'>
                My Account
              </button>
            </Link>
            <button
              onClick={handleLogout}
              className='custom-text text-xs px-5 py-2.5 font-semibold rounded-lg bg-cyan-400 text-white hover:bg-cyan-500 hover:cursor-pointer'
            >
              Logout
            </button>
          </>
        ) : (
          <>
            <div className='flex space-x-2'>
              <AuthModal initialMode='login'>
                <button className='custom-text text-xs px-5 py-2.5 font-semibold rounded-lg hover:bg-gray-100 text-gray-900 hover:cursor-pointer'>
                  Login
                </button>
              </AuthModal>

              <AuthModal initialMode='signup'>
                <button className='custom-text text-xs px-5 py-2.5 font-semibold rounded-lg bg-cyan-400 text-white hover:bg-cyan-500 hover:cursor-pointer'>
                  Sign Up
                </button>
              </AuthModal>
            </div>
          </>
        )}
      </div>
    </nav>
  );
}
