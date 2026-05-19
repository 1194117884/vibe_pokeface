"use client";

export interface ToastItem {
  id: number;
  text: string;
  seat: number;
}

interface GameNotificationsProps {
  toasts: ToastItem[];
}

export function GameNotifications({ toasts }: GameNotificationsProps) {
  if (toasts.length === 0) return null;

  return (
    <div className="fixed top-24 left-1/2 -translate-x-1/2 z-50 flex flex-col items-center gap-2 pointer-events-none">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className="animate-toast-in bg-black/80 backdrop-blur-md text-white text-base font-bold px-5 py-2.5 rounded-full shadow-lg whitespace-nowrap border border-white/10"
        >
          {toast.text}
        </div>
      ))}
    </div>
  );
}
