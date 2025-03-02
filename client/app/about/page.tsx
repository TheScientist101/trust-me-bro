import { title } from "@/components/primitives";

export default function AboutPage() {
    return (
        <div className="flex flex-col w-full justify-center items-center gap-4">
            <h1 className={title()}>About</h1>
            <p>Come on bro, you really want to know about me? Just trust me bro.</p>
        </div>
    );
}