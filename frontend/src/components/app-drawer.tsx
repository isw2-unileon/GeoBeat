import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Combobox, ComboboxEmpty, ComboboxInput, ComboboxList, ComboboxItem, ComboboxContent } from "@/components/ui/combobox"
import { modes, genres } from "@/data/placeholder-data"
import { useState } from "react"

type Props = {
  country: string;
};


export function AppDrawer({country}: Props) {
    const [isOpen, setIsOpen] = useState<boolean>(false);

    return (
        <Drawer open={isOpen} onOpenChange={setIsOpen}>
            <DrawerTrigger className="absolute top-10 right-4">
                <Button className="bg-white/80 text-black">Menu</Button>
            </DrawerTrigger>
            <DrawerContent>
                <DrawerHeader>
                    <DrawerTitle className="text-xl">GeoBeat</DrawerTitle>
                    <DrawerDescription>The not so hit music genre guessing game</DrawerDescription>
                </DrawerHeader>
                <div className="text-center mb-4">
                    <h1 className="text-base">Mode selection</h1>
                    <div className="max-w-xs mx-auto">
                        <Select onValueChange={(value) => {
                                console.log(value);
                                setIsOpen(false)
                            }}
                            defaultValue={modes[0]}
                        >
                            <SelectTrigger className="w-full">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectGroup>
                                    {modes.map((mode) => (
                                        <SelectItem key={mode} value={mode}>
                                        {mode}
                                        </SelectItem>
                                    ))}
                                </SelectGroup>
                            </SelectContent>
                        </Select>
                    </div>
                </div>
                <div className="text-center mb-4">
                    <h1 className="text-base">What is the most popular genre of?</h1>
                    <label>{country}</label>
                    <div className="max-w-xs mx-auto">
                        <Combobox items={genres}>
                            <ComboboxInput placeholder="Select a genre" />
                            <ComboboxContent>
                                <ComboboxEmpty>No genres available</ComboboxEmpty>
                                <ComboboxList>
                                    {(item: string) => (
                                        <ComboboxItem key={item} value={item}>
                                        {item}
                                        </ComboboxItem>
                                    )}
                                </ComboboxList>
                            </ComboboxContent>
                        </Combobox>
                    </div>
                </div>
                <DrawerFooter>
                    <DrawerClose>
                        <Button variant="outline">Close</Button>
                    </DrawerClose>
                </DrawerFooter>
            </DrawerContent>
        </Drawer>
    )
}