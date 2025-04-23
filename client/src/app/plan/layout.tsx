import { Navbar } from '@/components/ui/navbar';
import '../globals.css';

export default function PlanLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className='min-h-screen'>
      <Navbar />
      <div className='px-4 pt-4'>{children}</div>
    </div>
  );
}
