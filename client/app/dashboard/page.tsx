"use client";
import React, { useEffect } from "react";
import { title } from "@/components/primitives";
import { siteConfig } from "@/config/site";

export default function DashboardPage() {
    const [balance, setBalance] = React.useState<number>(0);
    const [firstName, setFirstName] = React.useState<string>("");
    const [lastName, setLastName] = React.useState<string>("");
    const [email, setEmail] = React.useState<string>("");
    const [accessToken, setAccessToken] = React.useState<string>("");
    const [Transactions, setTransactions] = React.useState<string[]>([]);

    useEffect(() => {
        setAccessToken(localStorage.getItem("access_token") || "");
    }, []);

    useEffect(() => {
        fetch(new URL("/userdata", siteConfig.links.server).href, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                access_token: accessToken,
            }),
        }).then((res) => res.text())
            .then((text) => JSON.parse(text))
            .then((data) => {
                setBalance(data.balance);
                setFirstName(data.first_name);
                setLastName(data.last_name);
                setEmail(data.email);
            });

        const n = 0; // Number of transactions
        const generatedTransactions = Array.from({ length: n }, (_, i) => `Transaction ${i + 1}`);
        setTransactions(generatedTransactions);
    }, [accessToken]);

    return (
        <div className="flex flex-col w-full items-center gap-4 py-8 md:py-10">
            <h1 className={title()}>Welcome, {firstName}!</h1>
            <div className="dashboard min-w-fit flex justify-between items-center bg-gray-400 p-5 mt-20 rounded-lg w-3/5 shadow-lg">
                <div className="dashboard-left flex-1 text-center text-2xl font-bold text-white">
                    Total Balance: ${balance}
                </div>
                <div className="dashboard-right flex-1 text-right text-lg">
                    <p>btw, your last name is: {lastName}</p>
                    <p>email: {email}</p>
                </div>
            </div>
            <div className="Transaction-list flex flex-wrap gap-5 w-3/5 mt-8 justify-center">
                {Transactions.length == 0 ? <p>No transactions currently available.</p> : Transactions.map((Transaction, index) => (
                    <div key={index} className="Transaction-card bg-gray-400 p-5 text-center rounded-lg shadow-lg text-xl font-bold text-yellow-400">
                        {Transaction}
                    </div>
                ))}
            </div>
            <style jsx>{`
                .nav-links {
                    list-style: none;
                    display: flex;
                    gap: 30px;
                    padding: 0;
                    margin: 0;
                }
                .nav-links li {
                    position: relative;
                }
                .nav-links li a {
                    text-decoration: none;
                    color: white;
                    font-size: 18px;
                    font-weight: bold;
                    padding: 10px;
                    transition: color 0.3s ease;
                }
                .nav-links li:hover a {
                    color: #007BFF;
                }
            `}</style>
        </div>
    );
}