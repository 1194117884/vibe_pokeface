import { ButtonHTMLAttributes, forwardRef } from "react";
import clsx from "clsx";

type ButtonVariant =
  | "primary"
  | "outlined"
  | "black-fill"
  | "dark-outlined"
  | "white-fill"
  | "outlined-light";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  fullWidth?: boolean;
}

const variantStyles: Record<ButtonVariant, string> = {
  primary: "bg-green-accent text-white border border-green-accent",
  outlined: "bg-transparent text-green-accent border border-green-accent",
  "black-fill": "bg-black text-white border border-black",
  "dark-outlined": "bg-transparent text-text-black border border-text-black",
  "white-fill": "bg-white text-green-accent border border-white",
  "outlined-light": "bg-transparent text-white border border-white",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  function Button({ variant = "primary", fullWidth, className, children, ...props }, ref) {
    return (
      <button
        ref={ref}
        className={clsx(
          "rounded-pill px-4 py-[7px] text-sm font-semibold tracking-tight",
          "transition-all duration-200 ease",
          "active:scale-[0.95]",
          "disabled:opacity-50 disabled:cursor-not-allowed",
          variantStyles[variant],
          fullWidth && "w-full",
          className
        )}
        {...props}
      >
        {children}
      </button>
    );
  }
);
