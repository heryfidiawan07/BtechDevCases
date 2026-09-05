import type { Metadata } from "next";
import { Fraunces, Source_Sans_3 } from "next/font/google";

import { AuthProvider } from "@/features/auth/AuthProvider";
import { OfflineBanner } from "@/components/OfflineBanner";

import "./globals.css";

const display = Fraunces({
  subsets: ["latin"],
  variable: "--font-display",
});

const sans = Source_Sans_3({
  subsets: ["latin"],
  variable: "--font-sans",
});

export const metadata: Metadata = {
  title: "Outpost",
  description: "Field-ready wallet with JWT auth and durable transfers",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`${display.variable} ${sans.variable} antialiased`}>
        <AuthProvider>
          <OfflineBanner />
          {children}
        </AuthProvider>
      </body>
    </html>
  );
}
