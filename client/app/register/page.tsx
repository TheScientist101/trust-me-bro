"use client";
import { Input } from "@heroui/input"
import { Form } from "@heroui/form";
import { Button } from "@heroui/button";
import { title } from "@/components/primitives";
import React, { FormEvent, useEffect } from "react";
import { siteConfig } from "@/config/site";

export default function RegisterPage() {
    const [firstName, setFirstName] = React.useState<string>("");
    const [lastName, setLastName] = React.useState<string>("");
    const [email, setEmail] = React.useState<string>("");
    const [password, setPassword] = React.useState<string>("");
    const [confirmPassword, setConfirmPassword] = React.useState<string>("");
    const [error, setError] = React.useState<string>("");
    const [success, setSuccess] = React.useState<boolean>(false);

    const onSubmit = (e: FormEvent) => {
        e.preventDefault();
        if (password !== confirmPassword) {
            return;
        }

        fetch(new URL("/register", siteConfig.links.server).href, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                first_name: firstName,
                last_name: lastName,
                email: email,
                password: password
            })
        }).then((res) => res.text())
            .then((text) => JSON.parse(text))
            .then((data) => {
                if (data.success) {
                    setSuccess(true);
                } else {
                    setError(data.error.error);
                }
            });
    }

    return (
        <div className="flex flex-col w-full justify-center items-center gap-4">
            <h1 className={title()}>Register</h1>
            <Form
                className="w-full max-w-md flex flex-col gap-4 items-center"
                onSubmit={onSubmit}
            >
                {success && <p className="text-white">Successfully registered! Check your email for a verification link. Be sure to check spam (it always goes there).</p>}
                {error && <p className="text-red-500">{error}</p>}
                <Input
                    isRequired
                    label="First Name"
                    labelPlacement="outside"
                    name="firstName"
                    placeholder="Enter your first name"
                    type="text"
                    onChange={(e) => setFirstName(e.target.value)}
                    value={firstName}
                />
                <Input
                    isRequired
                    label="Last Name"
                    labelPlacement="outside"
                    name="lastName"
                    placeholder="Enter your last name"
                    type="text"
                    onChange={(e) => setLastName(e.target.value)}
                    value={lastName}
                />
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
                <Input
                    isRequired
                    label="Confirm Password"
                    labelPlacement="outside"
                    name="confirmPassword"
                    placeholder="Confirm your password"
                    type="password"
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    value={confirmPassword}
                    validate={(value) => value === password ? undefined : "Passwords do not match"}
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