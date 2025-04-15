import { Hero } from './components/hero';
import { Navbar } from './components/navbar';

export default function Home() {
  return (
    <div className='min-h-[150vh] pt-24'>
      <Navbar />
      <Hero />
    </div>
  );
}
