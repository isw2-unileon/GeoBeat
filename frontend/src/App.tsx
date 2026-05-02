import { AppField } from './components/app-field';
import { AppDrawer } from './components/app-drawer';
import { AppDialog } from './components/app-dialog';
import { FreeModeMap } from './components/map/FreeModeMap';

import { useState } from 'react';

export default function App() {

  const [country, setCountry] = useState<string>('(Select a country by clicking on it)')

  return (
      <main className="relative min-h-screen flex flex-col">
        <DailyModeTitle />
        <AppDialog />
        <FreeModeMap country={country} setCountry={setCountry}/>
        {/* Desktop */}
        <div className='hidden md:block'>
          <AppField country={country} />
        </div>
        {/* Mobile */}
        <div className='md:hidden'>
          <AppDrawer country={country} />
        </div>
        <Attempts num={5}/>
        <div className='hidden'>
          <CorrectPopUp />
        </div>
      </main>
  )
}

function DailyModeTitle() {
  return (
  <h1 className="md:absolute md:top-6 md:left-14 md:text-5xl md:translate-x-0
                absolute top-2 left-1/2 -translate-x-1/2 text-outline
                text-2xl text-center text-blue-600 font-semibold font-[sans] animate-fade-in-down z-1">
    DAILY MODE
  </h1>
  )
}

function Attempts({num}: {num: number}) {

  return (
    <div className='bg-gray-100 rounded-sm absolute top-30 left-15 flex flex-row'>
      {[...Array(num)].map((_, i) => (
        <div key={i} className='bg-gray-200 w-8 h-8 m-2 rounded-sm' />
      ))}
    </div>
  )
}

function CorrectPopUp() {

  return(
    <label className='absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 animate-pop-fade text-6xl'>✅</label>
  )
}