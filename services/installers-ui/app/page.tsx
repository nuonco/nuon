import { Link, Video, Card } from "@/components";

export default async function Home({ searchParams }) {
  const queryString = new URLSearchParams(searchParams).toString();

  return (
    <>
      <header className="flex flex-auto flex-col gap-6 md:gap-12">
        <div className="text-center">
          <h1>
            <Link href="//nuon.co">Nuon</Link>
          </h1>
        </div>
      </header>
      <main>
        <p className="text-xl md:text-4xl text-center leading-relaxed">
          You should, in all likelihood, not be here.
        </p>
      </main>
    </>
  );
}
