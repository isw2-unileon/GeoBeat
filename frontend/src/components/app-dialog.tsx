import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import googleLogo from "@/graphics/google-icon.svg";
import { User } from "lucide-react"
import { FieldGroup, Field, FieldLabel, FieldSeparator, FieldDescription } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button"
import { emailLogin, emailRegister, googleLogin } from "@/services/auth"
import { useState } from "react";

export function AppDialog() {

    const [mode, setMode] = useState("login");

    return (
        <Dialog>
            <DialogTrigger asChild className="absolute md:top-8 md:right-82 top-20 right-4 z-1">
                <Button className="bg-white/80 text-black" size={"icon-lg"}><User /></Button>
            </DialogTrigger>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle className="text-lg">Select a method</DialogTitle>
                </DialogHeader>
                <form onSubmit={mode === "login" ? emailLogin : emailRegister}>
                    <FieldGroup>
                        <Field>
                            <FieldLabel> {mode === "login" ? "Login through mail" : "Create account with email"} </FieldLabel>
                        </Field>
                        {mode === "register" && (
                            <Field>
                                <FieldLabel> User name </FieldLabel>
                                <Input name="input-username" placeholder="Geouser" />
                            </Field>
                        )}
                        <Field>
                            <FieldLabel> Email </FieldLabel>
                            <Input name="input-email" placeholder="mail@example.com" />
                        </Field>
                        <Field>
                            <FieldLabel> Password </FieldLabel>
                            <Input name="input-password" placeholder="password" />
                        </Field>
                        <Field>
                            <Button variant={"default"} type="submit">
                                Submit
                            </Button>
                            {mode === "login" ? (
                                <FieldDescription>
                                    Don't have an account?
                                    <Button type="button" variant={"link"} className="ml-1 p-0 h-auto leading-none" onClick={() => setMode("register")}>
                                        Register
                                    </Button>
                                </FieldDescription>
                            ) : (
                                <FieldDescription>
                                    Already have an account?
                                    <Button type="button" variant={"link"} className="ml-1 p-0 h-auto leading-none" onClick={() => setMode("login")}>
                                        Log in
                                    </Button>
                                </FieldDescription>
                            )}
                        </Field>
                        <FieldSeparator />
                        <Field>
                            <FieldLabel className="text-base" > Login trough third party </FieldLabel>
                            <Button variant={"outline"} onClick={googleLogin}>
                                Google <img src={googleLogo} alt="Google logo" width={12}/>
                            </Button>
                        </Field>
                    </FieldGroup>
                </form>
            </DialogContent>
        </Dialog>
    )
}