import { redirect } from "next/navigation";

// Root page — redirect to dashboard; login handled by middleware
export default function Home() {
  redirect("/dashboard");
}
