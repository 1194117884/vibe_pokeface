"use client";

import { InputHTMLAttributes, forwardRef, useState } from "react";
import clsx from "clsx";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  function Input({ label, error, className = "", value, onChange, ...props }, ref) {
    const [focused, setFocused] = useState(false);
    const hasValue = value !== undefined && value !== "";

    return (
      <div className="relative">
        <div
          className={clsx(
            className,
            "border rounded-[4px] px-3 pt-5 pb-3 lg:pb-2 transition-all duration-200",
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
                ? "top-1 text-xs font-bold uppercase tracking-wide"
                : "top-1/2 -translate-y-1/2 text-base",
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
            className="w-full bg-transparent outline-none text-base text-text-black"
            onFocus={() => setFocused(true)}
            onBlur={() => setFocused(false)}
            value={value}
            onChange={onChange}
            {...props}
          />
        </div>
        {error && (
          <p className="mt-1 text-xs text-red-error">{error}</p>
        )}
      </div>
    );
  }
);
