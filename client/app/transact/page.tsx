"use client";
import { title } from "@/components/primitives";
import { siteConfig } from "@/config/site";
import { Button } from "@heroui/button";
import { Form } from "@heroui/form";
import { Input } from "@heroui/input";
import { useState } from "react";

export default function Transact() {
    const [recipientEmail, setRecipientEmail] = useState("");
    const [amount, setAmount] = useState("");

    const handleSubmit = (event: React.FormEvent) => {
        event.preventDefault();
        fetch(new URL("/send", siteConfig.links.server).href, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                to: recipientEmail,
                amount: parseFloat(amount),
                access_token: localStorage.getItem("access_token")
            })
        }).then((res) => res.text())
            .then((text) => JSON.parse(text))
            .then((data) => {
                if (data.success) {
                    alert("Money sent successfully!");
                } else {
                    console.log(data.error);
                }
            });
    };

    return (
        <div className="flex flex-col items-center justify-center gap-4 py-8 md:py-10">
            <h1 className={title()}>Send Money!</h1>
            <Form onSubmit={handleSubmit}>
                <Input
                    placeholder="Recipient Account Email"
                    fullWidth
                    required
                    value={recipientEmail}
                    onChange={(e) => setRecipientEmail(e.target.value)}
                />
                <Input
                    placeholder="Amount"
                    type="number"
                    fullWidth
                    required
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                />
                <Button type="submit" color="primary">
                    Send
                </Button>
            </Form>
        </div>
    );
}