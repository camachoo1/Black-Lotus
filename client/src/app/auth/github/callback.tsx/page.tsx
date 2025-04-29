'use client';

import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';

export default function AuthCallback() {
  const router = useRouter();
  const searchParams = useSearchParams();

  useEffect(() => {
    // Get the return path if specified
    const returnTo = searchParams.get('returnTo') || '/';

    // Add a small delay to ensure cookies are processed
    const redirectTimer = setTimeout(() => {
      router.push(returnTo);
    }, 500);

    return () => clearTimeout(redirectTimer);
  }, [router, searchParams]);

  return (
    <div className='flex flex-col items-center justify-center min-h-screen'>
      <div className='text-center'>
        <h1 className='text-2xl font-bold mb-4'>
          Authentication Successful
        </h1>
        <p className='text-gray-600 mb-4'>
          Redirecting you back to the application...
        </p>
        <div className='animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500 mx-auto'></div>
      </div>
    </div>
  );
}
