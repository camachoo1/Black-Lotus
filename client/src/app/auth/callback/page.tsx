'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';

export default function AuthCallback() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Get the return path if specified
    const returnTo = searchParams.get('returnTo') || '/';

    // Check for errors
    const errorMsg = searchParams.get('error');
    if (errorMsg) {
      setError(errorMsg);
      return;
    }

    // Add a small delay to ensure cookies are processed
    const redirectTimer = setTimeout(() => {
      router.push(returnTo);
    }, 1000);

    return () => clearTimeout(redirectTimer);
  }, [router, searchParams]);

  if (error) {
    return (
      <div className='flex flex-col items-center justify-center min-h-screen p-4'>
        <div className='bg-red-50 border border-red-300 p-4 rounded-md'>
          <h2 className='text-xl text-red-800 font-bold'>
            Authentication Error
          </h2>
          <p className='text-red-700'>{error}</p>
          <button
            onClick={() => router.push('/')}
            className='mt-4 px-4 py-2 bg-red-600 text-white rounded-md'
          >
            Return Home
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className='flex flex-col items-center justify-center min-h-screen p-4'>
      <div className='animate-pulse text-center'>
        <h2 className='text-xl font-semibold mb-2'>
          Completing your authentication...
        </h2>
        <p className='text-gray-600'>
          You&aposll be redirected momentarily
        </p>
      </div>
    </div>
  );
}
