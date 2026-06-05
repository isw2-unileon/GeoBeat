import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer";
import { Button } from "@/components/ui/button";
import { GameStatus, STATUS } from "@/types/gameTypes";
import { ModeSelect } from "@/components/mode-select";
import { useHandleInverseGuess } from "@/hooks/handleInverseGuess";
import { getDataFromISO } from "@/lib/getCountryData";
import { useState } from "react";
type Props = {
  song: string;
  countryISO: string;
  mode: string;
  setMode: React.Dispatch<React.SetStateAction<string>>;
  inverseStatus: GameStatus;
  setInverseStatus: React.Dispatch<React.SetStateAction<GameStatus>>;
};

export function InverseDrawer({
  song,
  countryISO,
  mode,
  setMode,
  inverseStatus,
  setInverseStatus,
}: Props) {
  const handleInverseGuess = useHandleInverseGuess();
  const country = getDataFromISO(countryISO).name;
  const [isOpen, setIsOpen] = useState<boolean>(false);

  return (
    <Drawer open={isOpen} onOpenChange={setIsOpen} modal={false}>
      <DrawerTrigger className="absolute top-10 right-4" asChild>
        <Button className="bg-white/80 text-black">Menu</Button>
      </DrawerTrigger>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle className="text-xl">GeoBeat</DrawerTitle>
          <DrawerDescription className="max-w-xs mx-auto">
            In guess the country mode you are given a song and have to guess
            where that song is from, in this mode the are no hints!
          </DrawerDescription>
        </DrawerHeader>
        <div className="text-center mb-4">
          <h1 className="text-base">Mode selection</h1>
          <div className="max-w-xs mx-auto">
            <ModeSelect mode={mode} setMode={setMode} setIsOpen={setIsOpen} />
          </div>
        </div>
        <div className="text-center mb-4">
          <h1 className="text-base">¿Where is this song from?</h1>
          <label>{song}</label>
          <div className="max-w-xs mx-auto">
            <span className="text-sm">Selected country: {country}</span>
          </div>
        </div>
        <div className="w-full flex justify-center">
          {inverseStatus.status !== STATUS.WON &&
          inverseStatus.status !== STATUS.LOST ? (
            <Button
              type="button"
              className="w-1/3"
              onClick={() => {
                handleInverseGuess(countryISO, country, setInverseStatus);
                setIsOpen(false);
              }}
            >
              Guess
            </Button>
          ) : (
            <Button disabled className="w-1/3">
              Guess
            </Button>
          )}
        </div>
        <DrawerFooter>
          <DrawerClose asChild>
            <Button type="button" variant="outline" className="mx-auto">
              Close
            </Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}
