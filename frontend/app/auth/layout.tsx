import Link from "next/link";

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div
      className="min-h-screen flex flex-col items-center justify-center bg-cream px-4"
      style={{
        paddingTop: "var(--safe-area-top, 0px)",
        paddingBottom: "var(--safe-area-bottom, 0px)",
        paddingLeft: "var(--safe-area-left, 0px)",
        paddingRight: "var(--safe-area-right, 0px)",
      }}
    >
      <Link href="/" className="mb-6 lg:mb-8 flex items-center gap-2">
        <span className="text-3xl">🎴</span>
        <span className="text-xl font-bold text-starbucks tracking-tight">
          PokeFace
        </span>
      </Link>
      <div className="w-full max-w-md mx-0 lg:mx-4">{children}</div>
    </div>
  );
}
