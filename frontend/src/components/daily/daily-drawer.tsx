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
import { useState } from "react";
import { GameStatus, STATUS } from "@/types/gameTypes";
import { useHandleGuess } from "@/hooks/handleGuess";
import { GenreCombobox } from "../genre-combobox";
import { ModeSelect } from "../mode-select";
import { useNormalizeGenre } from "@/hooks/useNormalizeGenre";
type Props = {
  country: string;
  mode: string;
  setMode: React.Dispatch<React.SetStateAction<string>>;
  gameStatus: GameStatus;
  setGameStatus: React.Dispatch<React.SetStateAction<GameStatus>>;
};

export function DailyDrawer({
  country,
  setMode,
  mode,
  gameStatus,
  setGameStatus,
}: Props) {
  const [genre, setGenre] = useState<string | null>(null);
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const handleGuess = useHandleGuess();
  const normalizeGuess = useNormalizeGenre();

  return (
    <Drawer open={isOpen} onOpenChange={setIsOpen} modal={false}>
      <DrawerTrigger className="absolute top-10 right-4" asChild>
        <Button className="bg-white/80 text-black">Menu</Button>
      </DrawerTrigger>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle className="text-xl">GeoBeat</DrawerTitle>
          <DrawerDescription className="max-w-xs mx-auto">
            <strong className="block">{mode}</strong>
            In daily mode you are given a country and have to guess the most
            popular song for it. Every mistake gives a hint in form of a song
          </DrawerDescription>
        </DrawerHeader>
        <div className="text-center mb-4">
          <h1 className="text-base">Mode selection</h1>
          <div className="max-w-xs mx-auto">
            <ModeSelect mode={mode} setMode={setMode} setIsOpen={setIsOpen} />
          </div>
        </div>
        <div className="text-center mb-4">
          <h1 className="text-base">What is the most popular genre of?</h1>
          <label>{country}</label>
          <div className="max-w-xs mx-auto">
            <GenreCombobox genre={genre} setGenre={setGenre} />
          </div>
        </div>
        <div className="w-full flex justify-center">
          {gameStatus.status !== STATUS.WON &&
          gameStatus.status !== STATUS.LOST ? (
            <Button
              type="button"
              className="w-1/3"
              onClick={() => {
                const normalized_genre = normalizeGuess(genre);
                handleGuess(normalized_genre, genre, setGameStatus);
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
