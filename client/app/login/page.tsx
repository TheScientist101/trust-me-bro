"use client";
import { Input } from "@heroui/input"
import { Form } from "@heroui/form";
import { Button } from "@heroui/button";
import { title } from "@/components/primitives";
import React, { FormEvent, useEffect } from "react";
import { siteConfig } from "@/config/site";

export default function LoginPage() {
    const [email, setEmail] = React.useState<string>("");
    const [password, setPassword] = React.useState<string>("");
    const [error, setError] = React.useState<string>("");
    const [accessToken, setAccessToken] = React.useState<string>("");
    const [accessTokenExpiry, setAccessTokenExpiry] = React.useState<string>("");

    useEffect(() => {
        localStorage.setItem("access_token", accessToken);
        localStorage.setItem("access_token_expiry", accessTokenExpiry);
    }, [accessTokenExpiry]);

    const onSubmit = (e: FormEvent) => {
        e.preventDefault();
        fetch(new URL("/login", siteConfig.links.server).href, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                email: email,
                password: password
            })
        }).then((res) => res.text())
            .then((text) => JSON.parse(text))
            .then((data) => {
                if (data.success) {
                    setAccessToken(data.access_token);
                    setAccessTokenExpiry((new Date().getTime() + 1000 * 60 * 15).toString());
                    window.location.href = "/dashboard";
                } else {
                    setError(data.error.error);
                }
            });
    }

    return (
        <div className="flex flex-col w-full justify-center items-center gap-4">
            <h1 className={title()}>Login</h1>
            <Form
                className="w-full max-w-md flex flex-col gap-4 items-center"
                onSubmit={onSubmit}
            >
                {error && <p className="text-red-500">{error}</p>}
                <Input
                    isRequired
                    errorMessage="Please enter a valid email"
                    label="Email"
                    labelPlacement="outside"
                    name="email"
                    placeholder="Enter your email"
                    type="email"
                    onChange={(e) => setEmail(e.target.value)}
                    value={email}
                />
                <Input
                    isRequired
                    label="Password"
                    labelPlacement="outside"
                    name="password"
                    placeholder="Enter your password"
                    type="password"
                    onChange={(e) => setPassword(e.target.value)}
                    value={password}
                />
                <div className="flex gap-2">
                    <Button color="primary" type="submit">
                        Submit
                    </Button>
                </div>
            </Form>
        </div>
    );
}