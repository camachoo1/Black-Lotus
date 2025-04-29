'use client';

import { useRouter, useSearchParams } from 'next/navigation';

export default function AuthError() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const error =
    searchParams.get('message') || 'An unknown error occurred';

  return (
    <div className='flex flex-col items-center justify-center min-h-screen p-4'>
      <div className='bg-red-50 border border-red-300 p-4 rounded-md max-w-md'>
        <h2 className='text-xl text-red-800 font-bold'>
          Authentication Failed
        </h2>
        <p className='text-red-700 mt-2'>{error}</p>
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
