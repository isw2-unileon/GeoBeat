import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import googleLogo from "@/graphics/google-icon.svg";
import { User } from "lucide-react"
import { FieldGroup, Field, FieldLabel, FieldSeparator } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

export function AppDialog() {
    return (
        <Dialog>
            <DialogTrigger asChild className="absolute md:top-8 md:right-82 top-20 right-4 z-1">
                <Button className="bg-white/80 text-black" size={"icon-lg"}><User /></Button>
            </DialogTrigger>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle className="text-lg">Select a login method</DialogTitle>
                </DialogHeader>
                <FieldGroup>
                    <Field>
                        <FieldLabel> Login trough mail </FieldLabel>
                    </Field>
                    <Field>
                        <FieldLabel> Email </FieldLabel>
                        <Input id="input-email" placeholder="mail@example.com" />
                    </Field>
                    <Field>
                        <FieldLabel> Password </FieldLabel>
                        <Input id="input-password" placeholder="password" />
                    </Field>
                    <FieldSeparator />
                    <Field>
                        <FieldLabel className="text-base"> Login trough third party </FieldLabel>
                        <Button variant={"outline"}>Google <img src={googleLogo} alt="Google logo" width={12} /></Button>
                    </Field>
                </FieldGroup>
            </DialogContent>
        </Dialog>
    )
}