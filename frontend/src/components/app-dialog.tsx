import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import googleLogo from "@/graphics/google-icon.svg";
import { User } from "lucide-react";
import {
  FieldGroup,
  Field,
  FieldLabel,
  FieldSeparator,
  FieldDescription,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  emailLogin,
  emailRegister,
  googleLogin,
  logout,
} from "@/services/auth";
import { useState } from "react";

const MODES = {
  LOGIN: "login",
  REGISTER: "register",
} as const;

type Mode = (typeof MODES)[keyof typeof MODES];

export function AppDialog() {
  const [mode, setMode] = useState<Mode>(MODES.LOGIN);
  let logged = localStorage.getItem("token") != null;

  return (
    <Dialog>
      <DialogTrigger
        asChild
        className="absolute md:top-8 md:right-82 top-20 right-4 z-1"
      >
        <Button className="bg-white/80 text-black" size={"icon-lg"}>
          <User />
        </Button>
      </DialogTrigger>
      <DialogContent>
        {!logged && <NotLoggedContet mode={mode} setMode={setMode} />}
        {logged && <LoggedContent />}
      </DialogContent>
    </Dialog>
  );
}

type NotLoggedConterProps = {
  mode: Mode;
  setMode: React.Dispatch<React.SetStateAction<Mode>>;
};

function NotLoggedContet({ mode, setMode }: NotLoggedConterProps) {
  return (
    <>
      <DialogHeader>
        <DialogTitle className="text-lg">Select a method</DialogTitle>
      </DialogHeader>
      <form onSubmit={mode === MODES.LOGIN ? emailLogin : emailRegister}>
        <FieldGroup>
          <Field>
            <FieldLabel>
              {mode === MODES.LOGIN
                ? "Login through mail"
                : "Create account with email"}
            </FieldLabel>
          </Field>
          {mode === MODES.REGISTER && (
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
            {mode === MODES.LOGIN ? (
              <FieldDescription>
                Don't have an account?
                <Button
                  type="button"
                  variant={"link"}
                  className="ml-1 p-0 h-auto leading-none"
                  onClick={() => setMode(MODES.REGISTER)}
                >
                  Register
                </Button>
              </FieldDescription>
            ) : (
              <FieldDescription>
                Already have an account?
                <Button
                  type="button"
                  variant={"link"}
                  className="ml-1 p-0 h-auto leading-none"
                  onClick={() => setMode(MODES.LOGIN)}
                >
                  Log in
                </Button>
              </FieldDescription>
            )}
          </Field>
          <FieldSeparator />
          <Field>
            <FieldLabel className="text-base">
              {" "}
              Login trough third party{" "}
            </FieldLabel>
            <Button type="button" variant={"outline"} onClick={googleLogin}>
              Google <img src={googleLogo} alt="Google logo" width={12} />
            </Button>
          </Field>
        </FieldGroup>
      </form>
    </>
  );
}

function LoggedContent() {
  return (
    <>
      <DialogHeader>
        <DialogTitle className="text-lg">Alredy logged in</DialogTitle>
      </DialogHeader>
      <Button
        type="button"
        variant={"destructive"}
        onClick={(e) => logout(e, 1)}
      >
        Log out
      </Button>
    </>
  );
}
