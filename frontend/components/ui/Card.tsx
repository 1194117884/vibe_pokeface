import { HTMLAttributes, forwardRef } from "react";
import clsx from "clsx";

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  padding?: "sm" | "md" | "lg";
}

const paddingStyles = {
  sm: "p-space-3",
  md: "p-space-4",
  lg: "p-space-5",
};

export const Card = forwardRef<HTMLDivElement, CardProps>(
  function Card({ padding = "md", className, children, ...props }, ref) {
    return (
      <div
        ref={ref}
        className={clsx(
          "bg-white rounded-[12px] shadow-card",
          paddingStyles[padding],
          className
        )}
        {...props}
      >
        {children}
      </div>
    );
  }
);
