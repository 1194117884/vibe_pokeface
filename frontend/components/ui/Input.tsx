"use client";

import { InputHTMLAttributes, ReactNode, forwardRef, useState } from "react";
import clsx from "clsx";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  rightElement?: ReactNode;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  function Input({ label, error, className = "", value, onChange, rightElement, ...props }, ref) {
    const [focused, setFocused] = useState(false);
    const hasValue = value !== undefined && value !== "";

    return (
      <div className="relative">
        <div
            className={clsx(
              className,
            "min-h-16 border rounded-[8px] px-4 pt-6 pb-3 transition-all duration-200",
            error
              ? "border-red-error bg-red-50/5"
              : focused
                ? "border-green-accent"
                : "border-gray-300"
          )}
        >
          <label
            className={clsx(
              "absolute left-3 transition-all duration-200 pointer-events-none",
              focused || hasValue
                ? "top-1.5 text-sm font-bold uppercase tracking-wide"
                : "top-1/2 -translate-y-1/2 text-lg",
              error
                ? "text-red-error"
                : focused
                  ? "text-green-accent"
                  : "text-text-black-soft"
            )}
          >
            {label}
          </label>
          <input
            ref={ref}
            className={clsx(
              "w-full bg-transparent outline-none text-lg text-text-black",
              rightElement && "pr-10"
            )}
            onFocus={() => setFocused(true)}
            onBlur={() => setFocused(false)}
            value={value}
            onChange={onChange}
            {...props}
          />
          {rightElement && (
            <div className="absolute right-3 top-1/2 -translate-y-1/2">
              {rightElement}
            </div>
          )}
        </div>
        {error && (
          <p className="mt-1 text-base font-bold text-red-error">{error}</p>
        )}
      </div>
    );
  }
);
