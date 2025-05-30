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
import {
  emailValidation,
  passwordSignupValidation,
  passwordValidation,
} from '@/utils/validators';
import { useMutation } from '@tanstack/react-query';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { FaApple, FaGithub, FaGoogle } from 'react-icons/fa';

interface FormValues {
  name?: string;
  email: string;
  password: string;
}

// API base URL for direct backend communication
const API_BASE =
  process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export function AuthModal({
  children,
  initialMode = 'login',
}: {
  children: React.ReactNode;
  initialMode?: 'login' | 'signup';
}) {
  const [mode, setMode] = useState<'login' | 'signup'>(initialMode);
  const [open, setOpen] = useState(false);
  const { refreshUser, fetchWithCsrf } = useAuth();

  // Initialize React Hook Form
  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    clearErrors,
  } = useForm<FormValues>({
    defaultValues: {
      name: '',
      email: '',
      password: '',
    },
    mode: 'onBlur',
  });

  // Login/Signup mutation
  const authMutation = useMutation({
    mutationFn: async (data: FormValues) => {
      try {
        const endpoint =
          mode === 'login'
            ? `${API_BASE}/api/login`
            : `${API_BASE}/api/signup`;

        const response = await fetchWithCsrf(endpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(data),
        });

        // If response is not OK
        if (!response.ok) {
          // For authentication failures (401)
          if (response.status === 401) {
            throw new Error(
              'Invalid credentials. Please check your email and password and try again.'
            );
          }

          // For other error responses, try to parse the error message
          try {
            const errorData = await response.json();
            if (errorData.error) {
              throw new Error(errorData.error);
            }
          } catch (jsonError) {
            // If parsing fails, use a generic error message
            console.error(jsonError);
            throw new Error('An error occurred. Please try again.');
          }
        }

        // If response is OK, return the parsed JSON
        return await response.json();
      } catch (error) {
        // Rethrow errors
        if (error instanceof Error) {
          throw error;
        }
        throw new Error(
          'An unexpected error occurred. Please try again.'
        );
      }
    },
    onSuccess: () => {
      refreshUser();
      setOpen(false);
    },
    onError: () => {
      setValue('password', '', {
        shouldValidate: false,
        shouldDirty: false,
        shouldTouch: false,
      });
    },
  });

  const toggleMode = () => {
    setMode((prev) => (prev === 'login' ? 'signup' : 'login'));
  };

  // OAuth handler
  const handleOAuth = (provider: 'google' | 'github' | 'apple') => {
    clearErrors();
    const returnTo = encodeURIComponent(window.location.pathname);

    // We'll still use our frontend routes for the OAuth flow
    if (provider === 'github') {
      window.location.href = `/auth/github?returnTo=${returnTo}`;
    } else if (provider === 'google') {
      window.location.href = `/auth/google?returnTo=${returnTo}`;
    } else {
      window.location.href = `/auth/apple?returnTo=${returnTo}`;
    }
  };

  // Handles the submit logic
  const onSubmit = async (data: FormValues) => {
    try {
      authMutation.mutate(data);
    } catch (error) {
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
            {...register('email', emailValidation)}
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
            {...register(
              'password',
              mode === 'login'
                ? passwordValidation
                : passwordSignupValidation
            )}
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
            <div
              className='bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative mb-4'
              role='alert'
            >
              <span className='block sm:inline'>
                {authMutation.error.message}
              </span>
            </div>
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
            type='button'
            onClick={() => handleOAuth('google')}
            className='flex items-center justify-center gap-2 w-full px-4 py-2 rounded border border-gray-300 hover:bg-gray-100 transition hover:cursor-pointer'
          >
            <FaGoogle /> Continue with Google
          </button>
          <button
            type='button'
            onClick={() => handleOAuth('github')}
            className='flex items-center justify-center gap-2 w-full px-4 py-2 rounded border border-gray-300 hover:bg-gray-100 transition hover:cursor-pointer'
          >
            <FaGithub /> Continue with GitHub
          </button>
          <button
            type='button'
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
