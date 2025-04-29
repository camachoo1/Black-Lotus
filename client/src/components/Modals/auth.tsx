'use client';

import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import { useAuth } from '@/contexts/AuthContext';
import { useMutation } from '@tanstack/react-query';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { FaApple, FaGithub, FaGoogle } from 'react-icons/fa';

interface FormValues {
  name?: string;
  email: string;
  password: string;
}

export function AuthModal({
  children,
  initialMode = 'login',
}: {
  children: React.ReactNode;
  initialMode?: 'login' | 'signup';
}) {
  const [mode, setMode] = useState<'login' | 'signup'>(initialMode);
  const [open, setOpen] = useState(false);
  const { refreshUser } = useAuth();

  // Initialize React Hook Form
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({
    defaultValues: {
      name: '',
      email: '',
      password: '',
    },
  });

  // Login/Signup mutation
  const authMutation = useMutation({
    mutationFn: async (data: FormValues) => {
      const endpoint =
        mode === 'login' ? '/api/login' : '/api/signup';
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
        credentials: 'include',
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(
          error.error ||
            `${mode === 'login' ? 'Login' : 'Signup'} failed`
        );
      }

      return response.json();
    },
    onSuccess: () => {
      refreshUser();
      setOpen(false); // Close modal on success
    },
  });

  const toggleMode = () => {
    setMode((prev) => (prev === 'login' ? 'signup' : 'login'));
  };

  // THIS WILL CHANGE ONCE OAUTH IS SETUP
  const handleOAuth = (provider: 'google' | 'github' | 'apple') => {
    const returnTo = encodeURIComponent(window.location.pathname);
    
    if (provider === 'github') {
      window.location.href = `/auth/github?returnTo=${returnTo}`;
    } else {
      window.location.href = `/api/auth/${provider}?returnTo=${returnTo}`;
    }
  };

  // Handles the submit logic
  const onSubmit = async (data: FormValues) => {
    try {
      const endpoint =
        mode === 'login' ? '/api/login' : '/api/signup';

      const payload =
        mode === 'login'
          ? { email: data.email, password: data.password }
          : {
              name: data.name,
              email: data.email,
              password: data.password,
            };

      const response = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
        credentials: 'include',
      });

      if (!response.ok) {
        // Try to parse error response
        let errorMessage = `${
          mode === 'login' ? 'Login' : 'Signup'
        } Failed`;
        try {
          const errorData = await response.json();
          if (errorData.error) errorMessage = errorData.error;
        } catch {
          // Ignore JSON parsing errors and use default message
          errorMessage += `: ${response.statusText}`;
        }
        throw new Error(errorMessage);
      }

      // Try to parse the success response
      try {
        await response.json();
        // Update auth context
        refreshUser();
        // Close modal
        setOpen(false);
      } catch {
        // If no JSON but response was OK, still consider it a success
        refreshUser();
        setOpen(false);
      }
    } catch (error) {
      // Handle any errors
      console.error('Authentication error:', error);
    }
  };

  return (
    // DIALOG IMPORTS FROM SHADCN TO HANDLE MODAL DESIGN
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild onClick={() => setOpen(true)}>
        {children}
      </DialogTrigger>
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

        {/* EMAIL & PASSWORD LOGIN FORM */}
        <form
          className='flex flex-col gap-3 mt-4'
          onSubmit={handleSubmit(onSubmit)}
        >
          {/* NAME ONLY REQUIRED WHEN SIGNING UP */}
          {mode === 'signup' && (
            <>
              <input
                {...register('name', {
                  required: 'Name is required',
                })}
                type='text'
                placeholder='Name'
                className='border rounded px-3 py-2 text-sm'
              />
              {errors.name && (
                <span className='text-red-500 text-xs'>
                  {errors.name.message}
                </span>
              )}
            </>
          )}

          {/* INPUTS FOR FORM */}
          <input
            {...register('email', {
              required: 'Email is required',
              pattern: {
                value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                message: 'Invalid email address',
              },
            })}
            type='email'
            placeholder='Email'
            required
            className='border rounded px-3 py-2 text-sm'
          />
          {errors.email && (
            <span className='text-red-500 text-xs'>
              {errors.email.message}
            </span>
          )}

          <input
            {...register('password', {
              required: 'Password is required',
              minLength: {
                value: 8,
                message: 'Password must be at least 8 characters',
              },
            })}
            type='password'
            placeholder='Password'
            required
            className='border rounded px-3 py-2 text-sm'
          />
          {errors.password && (
            <span className='text-red-500 text-xs'>
              {errors.password.message}
            </span>
          )}

          <button
            type='submit'
            className='bg-cyan-500 hover:bg-cyan-600 text-white font-semibold py-2 rounded hover:cursor-pointer'
          >
            {authMutation.isPending
              ? 'Loading...'
              : mode === 'login'
              ? 'Login'
              : 'Sign Up'}
          </button>

          {authMutation.isError && (
            <span className='text-red-500 text-xs'>
              {authMutation.error.message}
            </span>
          )}
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
        <div className='flex items-center gap-4 my-2'>
          <div className='h-px flex-1 bg-gray-200' />
          <span className='text-xs text-gray-400 uppercase'>or</span>
          <div className='h-px flex-1 bg-gray-200' />
        </div>

        {/* OAUTH BUTTONS */}
        <div className='flex flex-col gap-2 mt-2'>
          <button
            onClick={() => handleOAuth('google')}
            className='flex items-center justify-center gap-2 w-full px-4 py-2 rounded border border-gray-300 hover:bg-gray-100 transition hover:cursor-pointer'
          >
            <FaGoogle /> Continue with Google
          </button>
          <button
            onClick={() => handleOAuth('github')}
            className='flex items-center justify-center gap-2 w-full px-4 py-2 rounded border border-gray-300 hover:bg-gray-100 transition hover:cursor-pointer'
          >
            <FaGithub /> Continue with GitHub
          </button>
          <button
            onClick={() => handleOAuth('apple')}
            className='flex items-center justify-center gap-2 w-full px-4 py-2 rounded border border-gray-300 hover:bg-gray-100 transition hover:cursor-pointer'
          >
            <FaApple /> Continue with Apple
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
