import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { modes } from "@/data/placeholder-data";

type Props = {
  mode: string;
  setMode: React.Dispatch<React.SetStateAction<string>>;
  setIsOpen?: React.Dispatch<React.SetStateAction<boolean>>;
};

export function ModeSelect({ mode, setMode, setIsOpen }: Props) {
  return (
    <Select
      defaultValue={mode}
      onValueChange={(value) => {
        setMode(value);
        setIsOpen?.(false);
      }}
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
  );
}
