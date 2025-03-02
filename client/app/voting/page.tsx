"use client";
import { title } from "@/components/primitives";
import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { Button } from "@heroui/button";
import { Image } from "@heroui/image";
import { siteConfig } from "@/config/site";

export default function VotingPage() {
    const [accessToken, setAccessToken] = useState<string>("");
    const [accessTokenExpiry, setAccessTokenExpiry] = useState<string>("");
    const [transactionId, setTransactionId] = useState<string>("");

    useEffect(() => {
        setAccessToken(localStorage.getItem("access_token") || "");
        setAccessTokenExpiry(localStorage.getItem("access_token_expiry") || "");
        let params = (new URL(window.location.href)).searchParams;
        setTransactionId(params.get("id") as string || "");
    }, []);

    const trust = () => {
        fetch(new URL("/vote", siteConfig.links.server), {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                access_token: accessToken,
                transaction_id: transactionId,
                trust: true
            })
        }).then((res) => res.text())
            .then((text) => JSON.parse(text))
            .then((data) => {
                console.log(data);
                window.location.href = "/dashboard";
            });
    };

    const cap = () => {
        fetch(new URL("/vote", siteConfig.links.server), {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                access_token: accessToken,
                transaction_id: transactionId,
                trust: false
            })
        }).then((res) => res.text())
            .then((text) => JSON.parse(text))
            .then((data) => {
                console.log(data);
                window.location.href = "/dashboard";
            });
    };

    return (
        <div className="flex flex-col items-center justify-center gap-4 py-8 md:py-10">
            <h1 className={title()}>Vote on a Project</h1>
            <Button onPressEnd={trust}>Trust me bro</Button>
            <Button onPressEnd={cap}>Cap</Button>
            <Image src="https://imgflip.com/s/meme/Two-Buttons.jpg" alt="rizz.jpg" />
        </div>
    );
}