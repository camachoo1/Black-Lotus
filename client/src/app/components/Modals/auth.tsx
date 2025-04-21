'use client';

import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import { useState } from 'react';
import { FaApple, FaGithub, FaGoogle } from 'react-icons/fa';

export function AuthModal({
  children,
  initialMode = 'login',
}: {
  children: React.ReactNode;
  initialMode?: 'login' | 'signup';
}) {
  const [mode, setMode] = useState<'login' | 'signup'>(initialMode);

  const toggleMode = () => {
    setMode((prev) => (prev === 'login' ? 'signup' : 'login'));
  };

  const handleOAuth = (provider: 'google' | 'github' | 'apple') => {
    window.location.href = `/api/auth/${provider}`;
  };

  return (
    <Dialog>
      <DialogTrigger asChild>{children}</DialogTrigger>

      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>
            {mode === 'login' ? 'Log In' : 'Sign Up'}
          </DialogTitle>
          <DialogDescription>
            {mode === 'login'
              ? 'Access your Black Lotus account.'
              : 'Create your Black Lotus account.'}
          </DialogDescription>
        </DialogHeader>

        <form className='flex flex-col gap-3 mt-4'>
          <input
            type='email'
            placeholder='Email'
            required
            className='border rounded px-3 py-2 text-sm'
          />
          <input
            type='password'
            placeholder='Password'
            required
            className='border rounded px-3 py-2 text-sm'
          />
          <button
            type='submit'
            className='bg-cyan-500 hover:bg-cyan-600 text-white font-semibold py-2 rounded'
          >
            {mode === 'login' ? 'Login' : 'Sign Up'}
          </button>
        </form>

        <p className='text-xs text-center mt-2 text-gray-500'>
          {mode === 'login'
            ? "Don't have an account?"
            : 'Already have one?'}{' '}
          <button
            onClick={toggleMode}
            className='text-cyan-500 hover:underline ml-1'
          >
            {mode === 'login' ? 'Sign up' : 'Log in'}
          </button>
        </p>

        {/* Divider */}
        <div className='flex items-center gap-4 my-4'>
          <div className='h-px flex-1 bg-gray-200' />
          <span className='text-xs text-gray-400 uppercase'>or</span>
          <div className='h-px flex-1 bg-gray-200' />
        </div>

        {/* OAUTH BUTTONS */}
        <div className='flex flex-col gap-2 mt-4'>
          <button
            onClick={() => handleOAuth('google')}
            className='flex items-center justify-center gap-2 w-full px-4 py-2 rounded border border-gray-300 hover:bg-gray-50 transition'
          >
            <FaGoogle /> Continue with Google
          </button>
          <button
            onClick={() => handleOAuth('github')}
            className='flex items-center justify-center gap-2 w-full px-4 py-2 rounded border border-gray-300 hover:bg-gray-50 transition'
          >
            <FaGithub /> Continue with GitHub
          </button>
          <button
            onClick={() => handleOAuth('apple')}
            className='flex items-center justify-center gap-2 w-full px-4 py-2 rounded border border-gray-300 hover:bg-gray-50 transition'
          >
            <FaApple /> Continue with Apple
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
