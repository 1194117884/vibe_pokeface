import { AdminSidebar } from "@/components/ui/AdminSidebar";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen">
      <AdminSidebar />
      <main className="flex-1 bg-cream p-8">
        {children}
      </main>
    </div>
  );
}
