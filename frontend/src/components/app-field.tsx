import {
    FieldLegend,
    FieldSet,
    FieldDescription,
    FieldGroup,
    Field,
    FieldLabel,
    FieldSeparator
} from "@/components/ui/field"
import { Combobox, ComboboxEmpty, ComboboxInput, ComboboxList, ComboboxItem, ComboboxContent } from "@/components/ui/combobox"
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

type Props = {
  country: string;
};

const modes = ["Daily Mode", "WIP 1", "WIP 2"]
const genres = ["Rock", "Pop", "Country", "Jazz", "Rap", "Hip-Hop", "Classic"]

export function AppField({country}: Props) {
    return (
        <FieldSet className="absolute top-4 right-4 p-4 max-w-xs bg-white/80 rounded-md animate-fade-in-left">
            <FieldLegend className="!text-2xl bg-white rounded-md px-2"> GeoBeat </FieldLegend>
            <FieldDescription>The not so hit music genre guessing game</FieldDescription>
            <FieldGroup>
                <FieldSeparator />
                <Field>
                    <FieldLabel className="text-1xl">Mode selection</FieldLabel>
                    <Select>
                        <SelectTrigger className="w-full">
                            <SelectValue placeholder="Select a game mode" />
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
                </Field>
                <FieldSeparator />
                <Field>
                    <FieldLabel className="text-1xl">¿What is the most popular genre of?</FieldLabel>
                    <FieldLabel>{country}</FieldLabel>
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
                </Field>
            </FieldGroup>
        </FieldSet>
    )
}