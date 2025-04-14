import Image from 'next/image';

export default function Home() {
  return (
    <nav className='flex justify-between'>
      {/* LEFT SIDE OF NAVBAR */}
      <div className='flex items-center'>
        <Image
          src='/filled.svg'
          alt='logo'
          height={100}
          width={100}
        />
        <h1 className='flex font-bold text-2xl'>Black Lotus</h1>
      </div>
      {/* RIGHT SIDE OF NAVBAR */}
      <div className='flex items-center'>
        <div className='flex space-x-3'>
          <button className=''>Login</button>
          <button className=''>Sign Up</button>
        </div>
      </div>
    </nav>
  );
}
