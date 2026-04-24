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

export function AppDialog() {
    return (
        <Dialog>
            <DialogTrigger asChild className="absolute top-8 right-82 z-1">
                <Button variant={"outline"} size={"icon-lg"}><User /></Button>
            </DialogTrigger>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Select a login method</DialogTitle>
                </DialogHeader>
                <Button variant={"outline"}>Google <img src={googleLogo} alt="Google logo" width={12} /></Button>
            </DialogContent>
        </Dialog>
    )
}