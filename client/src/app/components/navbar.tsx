import Image from 'next/image';

export function Navbar() {
  return (
    <nav className='flex justify-around'>
      {/* LEFT SIDE OF NAVBAR */}
      <div className='flex items-center'>
        <Image
          src='/filled.svg'
          alt='logo'
          height={75}
          width={75}
        />
        <h1 className='flex custom-text font-bold text-lg'>Black Lotus</h1>
      </div>
      {/* RIGHT SIDE OF NAVBAR */}
      <div className='flex items-center'>
        <div className='flex px-3 m-2 space-x-2'>
          <button className='custom-text text-gray-900 focus:outline-none hover:bg-gray-100 focus:ring-4 focus:ring-gray-100 font-semibold rounded-lg text-xs px-5 py-2.5 me-2 mb-2'>
            Login
          </button>
          <button className='custom-text focus:outline-none text-white bg-cyan-400 hover:bg-cyan-500 focus:ring-4 focus:ring-cyan-300 font-semibold rounded-lg text-xs px-5 py-2.5 me-2 mb-2 dark:focus:ring-cyan-900'>
            Sign Up
          </button>
        </div>
      </div>
    </nav>
  );
}
